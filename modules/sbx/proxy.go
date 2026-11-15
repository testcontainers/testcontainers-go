package sbx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	createRequestLimit  = 16 << 20
	updateRequestLimit  = createRequestLimit
	createResponseLimit = 1 << 20
	proxyCloseTimeout   = 30 * time.Second
)

type dockerProxyServer struct {
	runtime   sandboxRuntime
	listener  net.Listener
	server    *http.Server
	forward   *httputil.ReverseProxy
	transport *http.Transport
	ports     *portManager
	mounts    *bindMountManager
	host      string

	cancel context.CancelFunc

	connectionsMu sync.Mutex
	connections   map[net.Conn]http.ConnState
	serveErr      chan error
	closeOnce     sync.Once
	closed        chan struct{}
	closeErr      error
}

type proxyOperation interface {
	onResponse(*http.Response) error
	onTransportError(error) error
}

type proxyOperationContextKey struct{}

func newDockerProxy(ctx context.Context, runtime sandboxRuntime) (proxyServer, error) {
	if runtime.name == "" || runtime.run == nil || runtime.dockerURL == nil {
		return nil, errors.New("initialize sbx Docker proxy: incomplete sandbox runtime")
	}
	mounts, err := newBindMountManager(runtime)
	if err != nil {
		return nil, err
	}
	transport, err := newDockerTransport(runtime.dockerURL)
	if err != nil {
		return nil, errors.Join(err, mounts.close())
	}
	client := &http.Client{Transport: transport}
	ports, err := newPortManager(ctx, runtime, client)
	if err != nil {
		transport.CloseIdleConnections()
		return nil, errors.Join(err, mounts.close())
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), proxyCloseTimeout)
		cleanupErr := ports.close(cleanupCtx)
		cancel()
		transport.CloseIdleConnections()
		return nil, errors.Join(fmt.Errorf("listen for sbx Docker proxy: %w", err), cleanupErr, mounts.close())
	}

	proxyCtx, cancel := context.WithCancel(context.Background())
	p := &dockerProxyServer{
		runtime: runtime, transport: transport, ports: ports,
		mounts: mounts, host: "tcp://" + listener.Addr().String(),
		cancel: cancel, connections: make(map[net.Conn]http.ConnState),
		serveErr: make(chan error, 1), closed: make(chan struct{}),
	}
	trackedListener := &connectionTrackingListener{Listener: listener, onClose: p.forgetConnection}
	p.listener = trackedListener
	p.forward = &httputil.ReverseProxy{
		Rewrite: func(request *httputil.ProxyRequest) {
			rewriteDockerRequest(request, runtime.dockerURL)
		},
		Transport:     transport,
		FlushInterval: -1,
		ErrorLog:      log.New(io.Discard, "", 0),
		ModifyResponse: func(response *http.Response) error {
			operation, _ := response.Request.Context().Value(proxyOperationContextKey{}).(proxyOperation)
			if operation == nil {
				return nil
			}
			return operation.onResponse(response)
		},
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, proxyErr error) {
			operation, _ := request.Context().Value(proxyOperationContextKey{}).(proxyOperation)
			if operation != nil {
				proxyErr = errors.Join(proxyErr, operation.onTransportError(proxyErr))
			}
			writeDockerError(writer, http.StatusBadGateway, proxyErr)
		},
	}
	p.server = &http.Server{
		Handler: p,
		BaseContext: func(net.Listener) context.Context {
			return proxyCtx
		},
		ConnState: p.trackConnection,
	}
	go func() {
		err := p.server.Serve(trackedListener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		p.serveErr <- err
		close(p.serveErr)
	}()
	return p, nil
}

func (p *dockerProxyServer) Host() string { return p.host }

func (p *dockerProxyServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	route := parseContainerRoute(request.Method, request.URL.Path)
	switch route.kind {
	case routeCreate:
		operation, err := p.prepareCreate(request, route.apiPrefix)
		if err != nil {
			writeDockerError(writer, http.StatusBadRequest, err)
			return
		}
		request = request.WithContext(context.WithValue(request.Context(), proxyOperationContextKey{}, operation))
	case routeUpdate:
		if err := validateUpdateRestartPolicy(request); err != nil {
			writeDockerError(writer, http.StatusBadRequest, err)
			return
		}
	case routeStart, routeRestart:
		finish, err := p.ports.beginStart(request.Context(), route.containerRef)
		if err != nil {
			writeDockerError(writer, http.StatusBadGateway, err)
			return
		}
		operation := &lifecycleOperation{finish: finish, accepted: statusSuccessful}
		request = request.WithContext(context.WithValue(request.Context(), proxyOperationContextKey{}, operation))
	case routeStop:
		operation := &lifecycleOperation{finish: p.ports.beginStop(route.containerRef), accepted: func(status int) bool { return status == http.StatusNoContent }}
		request = request.WithContext(context.WithValue(request.Context(), proxyOperationContextKey{}, operation))
	case routeKill:
		operation := &lifecycleOperation{finish: p.ports.beginStop(route.containerRef), accepted: func(int) bool { return false }}
		request = request.WithContext(context.WithValue(request.Context(), proxyOperationContextKey{}, operation))
	case routeDelete:
		linkOnly := dockerBoolValue(request.URL.Query().Get("link"))
		operation := &lifecycleOperation{
			finish: p.ports.beginDestroy(route.containerRef),
			accepted: func(status int) bool {
				// A 404 for an alias is not proof that the known canonical ID
				// disappeared; finish inspects that ID before releasing its lease.
				return !linkOnly && status == http.StatusNoContent
			},
		}
		request = request.WithContext(context.WithValue(request.Context(), proxyOperationContextKey{}, operation))
	case routeRename, routePrune:
		operation := &reconcileOperation{manager: p.ports}
		request = request.WithContext(context.WithValue(request.Context(), proxyOperationContextKey{}, operation))
	}
	p.forward.ServeHTTP(writer, request)
}

func (p *dockerProxyServer) prepareCreate(request *http.Request, apiPrefix string) (*createOperation, error) {
	body, err := readBounded(request.Body, createRequestLimit)
	_ = request.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read Docker container create request: %w", err)
	}

	payload, err := decodeDockerJSONObject(body, "Docker container create request", "HostConfig", "Labels", "Image", "ExposedPorts")
	if err != nil {
		return nil, fmt.Errorf("decode Docker container create request: %w", err)
	}
	payload.canonicalize()
	host := newDockerJSONObject("Docker HostConfig", "NetworkMode", "RestartPolicy", "PortBindings", "PublishAllPorts", "Binds", "Mounts")
	hostRaw, hadHost := payload.get("HostConfig")
	if hadHost && string(hostRaw) != "null" {
		host, err = decodeDockerJSONObject(hostRaw, "Docker HostConfig", "NetworkMode", "RestartPolicy", "PortBindings", "PublishAllPorts", "Binds", "Mounts")
		if err != nil {
			return nil, fmt.Errorf("decode Docker HostConfig: %w", err)
		}
		host.canonicalize()
	}
	if err := validateNetworkMode(host); err != nil {
		return nil, err
	}
	if err := validateRestartPolicy(host, true); err != nil {
		return nil, err
	}
	if err := canonicalizeDockerMounts(host); err != nil {
		return nil, err
	}

	bindings := make(portBindings)
	rawBindings, hadBindings := host.get("PortBindings")
	if hadBindings && string(rawBindings) != "null" {
		if err := json.Unmarshal(rawBindings, &bindings); err != nil {
			return nil, fmt.Errorf("decode Docker port bindings: %w", err)
		}
	}
	publishAll := false
	if raw, ok := host.get("PublishAllPorts"); ok && string(raw) != "null" {
		if err := json.Unmarshal(raw, &publishAll); err != nil {
			return nil, fmt.Errorf("decode Docker PublishAllPorts: %w", err)
		}
	}
	if publishAll {
		exposed, err := p.allExposedPorts(request.Context(), apiPrefix, payload)
		if err != nil {
			return nil, err
		}
		for port := range exposed {
			if len(bindings[port]) == 0 {
				bindings[port] = []portBinding{{}}
			}
		}
		host.set("PublishAllPorts", json.RawMessage("false"))
	}
	if err := validatePortBindings(bindings); err != nil {
		return nil, err
	}
	if err := rejectLeaseLabelCollision(payload); err != nil {
		return nil, err
	}
	if err := p.mounts.validate(host.values()); err != nil {
		return nil, err
	}

	prepared, err := p.ports.prepareCreate(request.Context(), bindings)
	if err != nil {
		return nil, err
	}
	operation := &createOperation{manager: p.ports, token: prepared.Token, containerName: request.URL.Query().Get("name")}
	abort := func(cause error) (*createOperation, error) {
		cleanupCtx, cancel := freshProxyContext()
		cleanupErr := p.ports.abortCreate(cleanupCtx, prepared.Token, false)
		cancel()
		return nil, errors.Join(cause, cleanupErr)
	}
	if err := p.mounts.prepare(request.Context(), host.values()); err != nil {
		return abort(err)
	}

	if hadBindings || publishAll || len(prepared.Bindings) > 0 {
		raw, marshalErr := json.Marshal(prepared.Bindings)
		if marshalErr != nil {
			return abort(fmt.Errorf("encode Docker port bindings: %w", marshalErr))
		}
		host.set("PortBindings", raw)
	}
	if len(host.fields) > 0 || hadHost {
		raw, marshalErr := json.Marshal(host)
		if marshalErr != nil {
			return abort(fmt.Errorf("encode Docker HostConfig: %w", marshalErr))
		}
		payload.set("HostConfig", raw)
	}
	if err := injectLeaseLabel(payload, prepared.Token); err != nil {
		return abort(err)
	}
	rewritten, err := json.Marshal(payload)
	if err != nil {
		return abort(fmt.Errorf("encode Docker container create request: %w", err))
	}
	request.Body = io.NopCloser(bytes.NewReader(rewritten))
	request.ContentLength = int64(len(rewritten))
	request.GetBody = nil
	request.TransferEncoding = nil
	request.Header.Del("Content-Length")
	return operation, nil
}

func (p *dockerProxyServer) allExposedPorts(ctx context.Context, apiPrefix string, payload *dockerJSONObject) (map[string]json.RawMessage, error) {
	exposed := make(map[string]json.RawMessage)
	if raw, ok := payload.get("ExposedPorts"); ok && string(raw) != "null" {
		if err := json.Unmarshal(raw, &exposed); err != nil {
			return nil, fmt.Errorf("decode Docker ExposedPorts: %w", err)
		}
	}
	var image string
	imageRaw, _ := payload.get("Image")
	if err := json.Unmarshal(imageRaw, &image); err != nil || image == "" {
		if err == nil {
			err = errors.New("image is empty")
		}
		return nil, fmt.Errorf("resolve image for PublishAllPorts: %w", err)
	}

	target := *p.runtime.dockerURL
	target.Path = apiPrefix + "/images/" + image + "/json"
	target.RawPath = ""
	target.RawQuery = ""
	inspectCtx, cancel := boundedContext(ctx, portOperationTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(inspectCtx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("make Docker image inspect request: %w", err)
	}
	response, err := p.transport.RoundTrip(request)
	if err != nil {
		return nil, fmt.Errorf("inspect Docker image %q for PublishAllPorts: %w", image, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := readBounded(response.Body, dockerResponseLimit)
		return nil, fmt.Errorf("inspect Docker image %q for PublishAllPorts: HTTP %d: %s", image, response.StatusCode, strings.TrimSpace(string(body)))
	}
	var inspect struct {
		Config struct {
			ExposedPorts map[string]json.RawMessage `json:"ExposedPorts"`
		} `json:"Config"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, dockerResponseLimit)).Decode(&inspect); err != nil {
		return nil, fmt.Errorf("decode Docker image %q inspection: %w", image, err)
	}
	for port, value := range inspect.Config.ExposedPorts {
		exposed[port] = value
	}
	return exposed, nil
}

func (p *dockerProxyServer) Close(ctx context.Context) error {
	p.closeOnce.Do(func() { go p.close() })
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.closed:
		return p.closeErr
	}
}

func (p *dockerProxyServer) close() {
	defer close(p.closed)
	p.cancel()
	var errs []error
	if err := p.server.Close(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errs = append(errs, err)
	}
	p.connectionsMu.Lock()
	connections := make([]net.Conn, 0, len(p.connections))
	for connection := range p.connections {
		connections = append(connections, connection)
	}
	p.connections = make(map[net.Conn]http.ConnState)
	p.connectionsMu.Unlock()
	for _, connection := range connections {
		if err := connection.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			errs = append(errs, err)
		}
	}
	managerCtx, cancel := context.WithTimeout(context.Background(), proxyCloseTimeout)
	errs = append(errs, p.ports.close(managerCtx))
	cancel()
	errs = append(errs, p.mounts.close())
	p.transport.CloseIdleConnections()
	if serveErr := <-p.serveErr; serveErr != nil {
		errs = append(errs, serveErr)
	}
	p.closeErr = errors.Join(errs...)
}

func (p *dockerProxyServer) trackConnection(connection net.Conn, state http.ConnState) {
	p.connectionsMu.Lock()
	defer p.connectionsMu.Unlock()
	if state == http.StateClosed {
		delete(p.connections, connection)
		return
	}
	// Hijacked connections remain tracked because net/http no longer owns or
	// reports their final state; Close must terminate them explicitly.
	p.connections[connection] = state
}

func (p *dockerProxyServer) forgetConnection(connection net.Conn) {
	p.connectionsMu.Lock()
	delete(p.connections, connection)
	p.connectionsMu.Unlock()
}

type connectionTrackingListener struct {
	net.Listener
	onClose func(net.Conn)
}

func (l *connectionTrackingListener) Accept() (net.Conn, error) {
	connection, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	tracked := &connectionTrackingConn{Conn: connection}
	if l.onClose != nil {
		tracked.onClose = func() { l.onClose(tracked) }
	}
	return tracked, nil
}

type connectionTrackingConn struct {
	net.Conn
	once    sync.Once
	onClose func()
}

func (c *connectionTrackingConn) Close() error {
	err := c.Conn.Close()
	if c.onClose != nil {
		c.once.Do(c.onClose)
	}
	return err
}

func (c *connectionTrackingConn) CloseRead() error {
	closer, ok := c.Conn.(interface{ CloseRead() error })
	if !ok {
		return errors.New("connection does not support closing its read side")
	}
	return closer.CloseRead()
}

func (c *connectionTrackingConn) CloseWrite() error {
	closer, ok := c.Conn.(interface{ CloseWrite() error })
	if !ok {
		return errors.New("connection does not support closing its write side")
	}
	return closer.CloseWrite()
}

type createOperation struct {
	manager       *portManager
	token         string
	containerName string
	once          sync.Once
	err           error
}

func (o *createOperation) onResponse(response *http.Response) error {
	o.once.Do(func() {
		switch {
		case response.StatusCode == http.StatusCreated:
			body, err := readBounded(response.Body, createResponseLimit)
			if err != nil {
				o.err = o.abort(true, fmt.Errorf("read Docker create response: %w", err))
				return
			}
			_ = response.Body.Close()
			response.Body = io.NopCloser(bytes.NewReader(body))
			response.ContentLength = int64(len(body))
			var result struct {
				ID string `json:"Id"`
			}
			if err := json.Unmarshal(body, &result); err != nil || result.ID == "" {
				if err == nil {
					err = errors.New("response has no container ID")
				}
				o.err = o.abort(true, fmt.Errorf("decode successful Docker create response: %w", err))
				return
			}
			if err := o.manager.bindCreate(o.token, result.ID, o.containerName); err != nil {
				o.err = o.abort(true, err)
			}
		case response.StatusCode >= 400 && response.StatusCode < 500:
			// Docker conclusively rejected the request. Keep its response intact;
			// cleanup errors remain recoverable by manager reconciliation/close.
			_ = o.abort(false, nil)
		default:
			// A server-side response cannot prove whether create committed.
			_ = o.abort(true, nil)
		}
	})
	return o.err
}

func (o *createOperation) onTransportError(_ error) error {
	o.once.Do(func() { o.err = o.abort(true, nil) })
	return o.err
}

func (o *createOperation) abort(ambiguous bool, cause error) error {
	ctx, cancel := freshProxyContext()
	defer cancel()
	return errors.Join(cause, o.manager.abortCreate(ctx, o.token, ambiguous))
}

type lifecycleOperation struct {
	finish   func(context.Context, bool) error
	accepted func(int) bool
	once     sync.Once
	err      error
}

func (o *lifecycleOperation) onResponse(response *http.Response) error {
	o.once.Do(func() {
		ctx, cancel := freshProxyContext()
		defer cancel()
		o.err = o.finish(ctx, o.accepted(response.StatusCode))
		if !statusSuccessful(response.StatusCode) {
			// Preserve the daemon's original error response. Reconciliation and
			// final cleanup retain any state that could not be inspected now.
			o.err = nil
		}
	})
	return o.err
}

func (o *lifecycleOperation) onTransportError(_ error) error {
	o.once.Do(func() {
		ctx, cancel := freshProxyContext()
		defer cancel()
		o.err = o.finish(ctx, false)
	})
	return o.err
}

type reconcileOperation struct {
	manager *portManager
	once    sync.Once
	err     error
}

func (o *reconcileOperation) onResponse(response *http.Response) error {
	if !statusSuccessful(response.StatusCode) {
		return nil
	}
	o.once.Do(func() {
		ctx, cancel := freshProxyContext()
		defer cancel()
		o.err = o.manager.reconcile(ctx)
	})
	return o.err
}

func (o *reconcileOperation) onTransportError(error) error { return nil }

type containerRouteKind int

const (
	routeOther containerRouteKind = iota
	routeCreate
	routeUpdate
	routeStart
	routeRestart
	routeStop
	routeKill
	routeDelete
	routeRename
	routePrune
)

type containerRoute struct {
	kind         containerRouteKind
	apiPrefix    string
	containerRef string
}

func parseContainerRoute(method, path string) containerRoute {
	clean := strings.TrimPrefix(path, "/")
	parts := strings.Split(clean, "/")
	prefix := ""
	if len(parts) > 0 && dockerAPIVersion(parts[0]) {
		prefix = "/" + parts[0]
		parts = parts[1:]
	}
	if len(parts) < 2 || parts[0] != "containers" {
		return containerRoute{}
	}
	if method == http.MethodPost && len(parts) == 2 && parts[1] == "create" {
		return containerRoute{kind: routeCreate, apiPrefix: prefix}
	}
	if method == http.MethodPost && len(parts) == 2 && parts[1] == "prune" {
		return containerRoute{kind: routePrune, apiPrefix: prefix}
	}
	if len(parts) == 2 && method == http.MethodDelete && parts[1] != "" {
		return containerRoute{kind: routeDelete, apiPrefix: prefix, containerRef: parts[1]}
	}
	if len(parts) != 3 || method != http.MethodPost || parts[1] == "" {
		return containerRoute{}
	}
	kinds := map[string]containerRouteKind{
		"start": routeStart, "restart": routeRestart, "stop": routeStop,
		"kill": routeKill, "rename": routeRename, "update": routeUpdate,
	}
	return containerRoute{kind: kinds[parts[2]], apiPrefix: prefix, containerRef: parts[1]}
}

func dockerAPIVersion(segment string) bool {
	if len(segment) < 2 || segment[0] != 'v' {
		return false
	}
	digit := false
	for _, character := range segment[1:] {
		switch {
		case character >= '0' && character <= '9':
			digit = true
		case character == '.' && digit:
			digit = false
		default:
			return false
		}
	}
	return digit
}

func validateNetworkMode(host *dockerJSONObject) error {
	raw, ok := host.get("NetworkMode")
	if !ok || string(raw) == "null" {
		return nil
	}
	var mode string
	if err := json.Unmarshal(raw, &mode); err != nil {
		return fmt.Errorf("decode Docker NetworkMode: %w", err)
	}
	if mode == "host" || mode == "container" || strings.HasPrefix(mode, "container:") {
		return fmt.Errorf("sbx does not support Docker network mode %q because it bypasses managed published ports", mode)
	}
	return nil
}

func validateRestartPolicy(host *dockerJSONObject, canonicalize bool) error {
	raw, ok := host.get("RestartPolicy")
	if !ok || string(raw) == "null" {
		return nil
	}
	policy, err := decodeDockerJSONObject(raw, "Docker RestartPolicy", "Name")
	if err != nil {
		// Let Docker preserve its normal validation error for malformed policy
		// shapes. Only a valid, explicitly enabled name can bypass our gate.
		if isDuplicateDockerJSONMemberError(err) {
			return err
		}
		return nil
	}
	if canonicalize {
		policy.canonicalize()
		canonical, marshalErr := json.Marshal(policy)
		if marshalErr != nil {
			return fmt.Errorf("encode Docker RestartPolicy: %w", marshalErr)
		}
		host.set("RestartPolicy", canonical)
	}
	nameRaw, ok := policy.get("Name")
	if !ok || string(nameRaw) == "null" {
		return nil
	}
	var name string
	if err := json.Unmarshal(nameRaw, &name); err != nil {
		return nil
	}
	if name == "" || name == "no" {
		return nil
	}
	return fmt.Errorf("sbx does not support Docker restart policy %q because automatic restarts bypass managed published ports", name)
}

func validateUpdateRestartPolicy(request *http.Request) error {
	body, err := readBounded(request.Body, updateRequestLimit)
	_ = request.Body.Close()
	if err != nil {
		return fmt.Errorf("read Docker container update request: %w", err)
	}
	request.Body = io.NopCloser(bytes.NewReader(body))

	payload, err := decodeDockerJSONObject(body, "Docker container update request", "RestartPolicy")
	if err != nil {
		// Forward malformed requests unchanged so Docker remains authoritative
		// for ordinary API validation and its error response.
		if isDuplicateDockerJSONMemberError(err) && json.Valid(body) {
			return err
		}
		return nil
	}
	return validateRestartPolicy(payload, false)
}

func dockerBoolValue(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "0", "no", "false", "none":
		return false
	default:
		return true
	}
}

func validatePortBindings(bindings portBindings) error {
	for port, entries := range bindings {
		if _, err := parseContainerPort(port); err != nil {
			return fmt.Errorf("invalid Docker container port %q: %w", port, err)
		}
		_, protocol, _ := strings.Cut(port, "/")
		if protocol != "tcp" && protocol != "udp" {
			return fmt.Errorf("invalid Docker container port %q: sbx supports only TCP or UDP", port)
		}
		for _, binding := range entries {
			if err := validateHostIP(binding.HostIP); err != nil {
				return fmt.Errorf("container port %s: %w", port, err)
			}
			if _, _, err := parseHostPort(binding.HostPort); err != nil {
				return fmt.Errorf("container port %s: %w", port, err)
			}
		}
	}
	return nil
}

func rejectLeaseLabelCollision(payload *dockerJSONObject) error {
	raw, ok := payload.get("Labels")
	if !ok || string(raw) == "null" {
		return nil
	}
	var labels map[string]string
	if err := json.Unmarshal(raw, &labels); err != nil {
		return fmt.Errorf("decode Docker container labels: %w", err)
	}
	if _, exists := labels[portLeaseLabel]; exists {
		return fmt.Errorf("docker container label %q is reserved by the sbx runtime", portLeaseLabel)
	}
	return nil
}

func injectLeaseLabel(payload *dockerJSONObject, token string) error {
	labels := make(map[string]string)
	if raw, ok := payload.get("Labels"); ok && string(raw) != "null" {
		if err := json.Unmarshal(raw, &labels); err != nil {
			return fmt.Errorf("decode Docker container labels: %w", err)
		}
	}
	labels[portLeaseLabel] = token
	raw, err := json.Marshal(labels)
	if err != nil {
		return fmt.Errorf("encode Docker container labels: %w", err)
	}
	payload.set("Labels", raw)
	return nil
}

func rewriteDockerRequest(request *httputil.ProxyRequest, target *url.URL) {
	request.Out.URL.Scheme = target.Scheme
	request.Out.URL.Host = target.Host
	request.Out.URL.User = target.User
	request.Out.URL.Path = joinURLPath(target.Path, request.In.URL.Path)
	if target.Path == "" {
		request.Out.URL.RawPath = request.In.URL.RawPath
	} else {
		request.Out.URL.RawPath = ""
	}
	request.Out.URL.RawQuery = request.In.URL.RawQuery
	request.Out.Host = target.Host
}

func joinURLPath(base, path string) string {
	if base == "" {
		return path
	}
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(path, "/")
}

func statusSuccessful(status int) bool { return status >= 200 && status < 300 }

func freshProxyContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), portOperationTimeout)
}

func readBounded(reader io.Reader, limit int64) ([]byte, error) {
	limited := io.LimitReader(reader, limit+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("body exceeds %s bytes", strconv.FormatInt(limit, 10))
	}
	return data, nil
}

func writeDockerError(writer http.ResponseWriter, status int, err error) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]string{"message": err.Error()})
}
