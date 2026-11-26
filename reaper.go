package testcontainers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// Deprecated: it has been replaced by the internal core.LabelLang
	TestcontainerLabel = "org.testcontainers.golang"
	// Deprecated: it has been replaced by the internal core.LabelSessionID
	TestcontainerLabelSessionID = TestcontainerLabel + ".sessionId"
	// Deprecated: it has been replaced by the internal core.LabelReaper
	TestcontainerLabelIsReaper = TestcontainerLabel + ".reaper"
)

var (
	// Deprecated: it has been replaced by an internal value
	ReaperDefaultImage = config.ReaperDefaultImage

	// defaultReaperPort is the default port that the reaper listens on if not
	// overridden by the RYUK_PORT environment variable.
	defaultReaperPort = network.MustParsePort("8080/tcp")

	// errReaperNotFound is returned when no reaper container is found.
	errReaperNotFound = errors.New("reaper not found")

	// defaultReaperCreatedGracePeriod is how long a reaper container may stay
	// created but not started before it is treated as orphaned. It is shorter
	// than the lookup backoff, so the lookup that first sees an orphan can
	// still replace it.
	defaultReaperCreatedGracePeriod = 10 * time.Second

	// reaperStartTimeout is how long creating and starting a reaper container
	// may take, on top of its ready timeout, before the run gives up.
	reaperStartTimeout = 10 * time.Second

	// reaperCleanupTimeout bounds removing the container of a reaper run that
	// failed.
	reaperCleanupTimeout = 30 * time.Second

	// reaperReadyTimeout is how long each attempt to create or reuse a reaper
	// waits for it to acknowledge a connection, before the next attempt
	// recreates it. It leaves time for that within the spawner's retries.
	reaperReadyTimeout = 15 * time.Second

	// reaperConnectTimeout bounds connecting to a reaper and its handshake.
	reaperConnectTimeout = 5 * time.Second

	// backgroundLogger is what a reaper run that outlives its caller logs to.
	// It's captured at package initialisation, so it can't be a test logger.
	backgroundLogger = log.Default()

	// errReaperDisabled is returned if a reaper is requested but the
	// config has it disabled.
	errReaperDisabled = errors.New("reaper disabled")

	// spawner is the singleton instance of reaperSpawner.
	spawner = newReaperSpawner()

	// reaperAck is the expected response from the reaper container.
	reaperAck = []byte("ACK\n")

	// errReaperPortNotPublished is returned when a reaper container's port has
	// no host binding, so nothing can connect to it.
	errReaperPortNotPublished = errors.New("reaper port not published")
)

// ReaperProvider represents a provider for the reaper to run itself with
// The ContainerProvider interface should usually satisfy this as well, so it is pluggable
type ReaperProvider interface {
	RunContainer(ctx context.Context, req ContainerRequest) (Container, error)
	Config() TestcontainersConfig
}

// Deprecated: it's not possible to create a reaper any more. Compose module uses this method
// to create a reaper for the compose stack.
//
// # NewReaper creates a Reaper with a sessionID to identify containers and a provider to use
//
// The caller must call Connect at least once on the returned Reaper and use the returned
// result otherwise the reaper will be kept open until the process exits.
func NewReaper(ctx context.Context, sessionID string, provider ReaperProvider, _ string) (*Reaper, error) {
	reaper, err := spawner.reaper(ctx, sessionID, provider)
	if err != nil {
		return nil, fmt.Errorf("reaper: %w", err)
	}

	return reaper, nil
}

// reaperContainerNameFromSessionID returns the container name that uniquely
// identifies the container based on the session id.
func reaperContainerNameFromSessionID(sessionID string) string {
	// The session id is 64 characters, so we will not hit the limit of 128
	// characters for container names.
	return "reaper_" + sessionID
}

// reaperSpawner is a singleton that manages the reaper container.
type reaperSpawner struct {
	instance *Reaper
	mtx      sync.Mutex

	// createdGracePeriod is how long a reaper container may stay created but
	// not started before it is treated as orphaned.
	createdGracePeriod time.Duration

	// cleanupTimeout bounds removing the container of a reaper run that failed.
	cleanupTimeout time.Duration

	// detachedLogger is what a reaper run that outlives its caller logs to.
	detachedLogger log.Logger

	// readyTimeout is how long each attempt to create or reuse a reaper waits
	// for it to be ready.
	readyTimeout time.Duration

	// waitStrategy returns the wait strategy for a reaper container of
	// sessionID, giving up after readyTimeout, which keeps the connection it
	// was acknowledged on in conn.
	waitStrategy func(sessionID string, conn *reaperConn) wait.Strategy
}

// newReaperSpawner returns a reaperSpawner with the default settings.
func newReaperSpawner() *reaperSpawner {
	r := &reaperSpawner{
		createdGracePeriod: defaultReaperCreatedGracePeriod,
		cleanupTimeout:     reaperCleanupTimeout,
		detachedLogger:     backgroundLogger,
		readyTimeout:       reaperReadyTimeout,
	}
	r.waitStrategy = r.defaultWaitStrategy

	return r
}

// port returns the port that a new reaper should listen on.
func (r *reaperSpawner) port() network.Port {
	if port := os.Getenv("RYUK_PORT"); port != "" {
		natPort, err := network.ParsePort(port + "/tcp")
		if err != nil {
			panic(fmt.Sprintf("invalid RYUK_PORT value %q: %s", port, err))
		}
		return natPort
	}

	return defaultReaperPort
}

// backoff returns a backoff policy for the reaper spawner.
// It will take at most 20 seconds, doing each attempt every 100ms - 250ms.
func (r *reaperSpawner) backoff() *backoff.ExponentialBackOff {
	// We want random intervals between 100ms and 250ms for concurrent executions
	// to not be synchronized: it could be the case that multiple executions of this
	// function happen at the same time (specifically when called from a different test
	// process execution), and we want to avoid that they all try to find the reaper
	// container at the same time.
	b := &backoff.ExponentialBackOff{
		InitialInterval:     time.Millisecond * 100,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		// Adjust MaxInterval to compensate for randomization factor which can be added to
		// returned interval so we have a maximum of 250ms.
		MaxInterval:    time.Duration(float64(time.Millisecond*250) * backoff.DefaultRandomizationFactor),
		MaxElapsedTime: time.Second * 20,
		Stop:           backoff.Stop,
		Clock:          backoff.SystemClock,
	}
	b.Reset()

	return b
}

// cleanup terminates the reaper container if set.
func (r *reaperSpawner) cleanup() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.cleanupLocked()
}

// cleanupLocked terminates the reaper container if set.
// It must be called with the lock held.
func (r *reaperSpawner) cleanupLocked() error {
	if r.instance == nil {
		return nil
	}

	err := TerminateContainer(r.instance.container)
	r.instance.close()
	r.instance = nil

	return err
}

// lookupContainer returns the running reaper container for sessionID, waiting
// for one that may still become ready. It returns errReaperNotFound if there
// is no reaper container, or if the existing one will never become ready, in
// which case it is removed so a new one can be created.
func (r *reaperSpawner) lookupContainer(ctx context.Context, sessionID string, provider ReaperProvider) (*DockerContainer, error) {
	p, release, err := lookupProvider(ctx, provider)
	if err != nil {
		return nil, err
	}
	defer release()

	opts := client.ContainerListOptions{
		All: true,
		Filters: make(client.Filters).
			Add("label", fmt.Sprintf("%s=%s", core.LabelSessionID, sessionID)).
			Add("label", fmt.Sprintf("%s=%t", core.LabelReaper, true)).
			Add("label", fmt.Sprintf("%s=%t", core.LabelRyuk, true)).
			Add("name", reaperContainerNameFromSessionID(sessionID)),
	}

	// Use the underlying client, as DockerClient.Info caches its result.
	var system client.SystemAPIClient = p.client
	if dc, ok := p.client.(*DockerClient); ok {
		system = dc.Client
	}
	daemonNow := daemonClock(ctx, system)

	return backoff.RetryWithData(
		func() (*DockerContainer, error) {
			resp, err := p.client.ContainerList(ctx, opts)
			if err != nil {
				return nil, fmt.Errorf("container list: %w", err)
			}

			if len(resp.Items) == 0 {
				// No reaper container not found.
				return nil, backoff.Permanent(errReaperNotFound)
			}

			if len(resp.Items) > 1 {
				return nil, fmt.Errorf("found %d reaper containers for session ID %q", len(resp.Items), sessionID)
			}

			item := resp.Items[0]
			if item.State != container.StateRunning {
				if err := r.checkStale(item, daemonNow); err != nil {
					// It may still become ready, and other test processes
					// may depend on it, so retry rather than remove it.
					return nil, err
				}

				if err := removeStaleReaper(ctx, p.client, item.ID); err != nil {
					return nil, err
				}

				return nil, backoff.Permanent(errReaperNotFound)
			}

			ctr, err := p.ContainerFromType(ctx, item)
			if err != nil {
				return nil, fmt.Errorf("from docker: %w", err)
			}

			switch ctr.healthStatus {
			case "", container.Healthy, container.NoHealthcheck:
				return ctr, nil
			default:
				return nil, fmt.Errorf("container not healthy: %s", ctr.healthStatus)
			}
		},
		backoff.WithContext(r.backoff(), ctx),
	)
}

// lookupProvider returns provider if it's a *DockerProvider, otherwise a new
// DockerProvider for the Docker host in ctx, with a function that releases it.
func lookupProvider(ctx context.Context, provider ReaperProvider) (*DockerProvider, func(), error) {
	if p, ok := provider.(*DockerProvider); ok {
		return p, func() {}, nil
	}

	cli, err := NewDockerClientWithOpts(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("new client: %w", err)
	}

	p, err := NewDockerProvider()
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("new provider: %w", err), cli.Close())
	}

	p.SetClient(cli)

	return p, func() { _ = p.Close() }, nil
}

// checkStale returns nil if the reaper container is stale, so will never
// become ready, otherwise an error saying why it may still become ready, or
// why that can't be determined. daemonNow returns the daemon's current time.
func (r *reaperSpawner) checkStale(item container.Summary, daemonNow func() (time.Time, error)) error {
	switch item.State {
	case container.StateExited, container.StateDead, container.StateRemoving:
		// A stopped reaper never restarts.
		return nil
	case container.StateCreated:
		// A reaper is started straight after it is created, so one still not
		// started after the grace period was orphaned. Created is set by the
		// daemon's clock, so its age is measured against that.
		if item.Created <= 0 {
			return errors.New("container not running: state created, at an unknown time")
		}

		now, err := daemonNow()
		if err != nil {
			return fmt.Errorf("container not running: state created: daemon time: %w", err)
		}

		if age := now.Sub(time.Unix(item.Created, 0)); age <= r.createdGracePeriod {
			return fmt.Errorf("container not running: state created %s ago", age.Round(time.Second))
		}

		return nil
	default:
		// Restarting, paused or unknown.
		return fmt.Errorf("container not running: state %s", item.State)
	}
}

// daemonClock returns a function that returns the current time of the daemon
// cli is connected to. The returned function is not safe for concurrent use.
func daemonClock(ctx context.Context, cli client.SystemAPIClient) func() (time.Time, error) {
	var offset *time.Duration
	return func() (time.Time, error) {
		if offset == nil {
			// Unlike sync.OnceValues, an error isn't cached, so it's retried.
			fetched, err := daemonClockOffset(ctx, cli)
			if err != nil {
				return time.Time{}, err
			}

			offset = &fetched
		}

		return time.Now().Add(*offset), nil
	}
}

// daemonClockOffset returns how far the clock of the daemon cli is connected
// to is ahead of the client's.
func daemonClockOffset(ctx context.Context, cli client.SystemAPIClient) (time.Duration, error) {
	res, err := cli.Info(ctx, client.InfoOptions{})
	if err != nil {
		return 0, fmt.Errorf("info: %w", err)
	}

	return clockOffset(res.Info.SystemTime)
}

// clockOffset returns how far systemTime, the daemon's current time in
// RFC 3339 format, is ahead of the client's clock.
func clockOffset(systemTime string) (time.Duration, error) {
	now, err := time.Parse(time.RFC3339Nano, systemTime)
	if err != nil {
		return 0, fmt.Errorf("parse system time %q: %w", systemTime, err)
	}

	return time.Until(now), nil
}

// removeStaleReaper removes the stale reaper container id, unless it has
// started since it was found to be stale. A container that has already been
// removed, or is being removed, is not an error.
func removeStaleReaper(ctx context.Context, cli client.ContainerAPIClient, id string) error {
	// Not forced, so the daemon refuses to remove a reaper that has since
	// started, which the lookup's retry then finds.
	if _, err := cli.ContainerRemove(ctx, id, client.ContainerRemoveOptions{}); !isCleanupSafe(err) {
		return fmt.Errorf("remove stale container: %w", err)
	}

	return nil
}

// isRunning returns an error if the container is not running.
func (r *reaperSpawner) isRunning(ctx context.Context, ctr Container) error {
	state, err := ctr.State(ctx)
	if err != nil {
		return fmt.Errorf("container state: %w", err)
	}

	if !state.Running {
		// Use NotFound error to indicate the container is not running
		// and should be recreated.
		return errdefs.ErrNotFound.WithMessage("container state: " + string(state.Status))
	}

	return nil
}

// retryError returns a permanent error if the error is not considered retryable.
func (r *reaperSpawner) retryError(err error) error {
	var timeout interface {
		Timeout() bool
	}
	switch {
	case isCleanupSafe(err),
		errors.Is(err, errReaperPortNotPublished),
		createContainerFailDueToNameConflictRegex.MatchString(err.Error()),
		errors.Is(err, syscall.ECONNREFUSED),
		errors.Is(err, syscall.ECONNRESET),
		errors.Is(err, syscall.ECONNABORTED),
		errors.Is(err, syscall.ETIMEDOUT),
		errors.Is(err, io.EOF),
		errors.Is(err, io.ErrUnexpectedEOF),
		errors.Is(err, os.ErrDeadlineExceeded),
		errors.As(err, &timeout) && timeout.Timeout(),
		errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		// Retryable error.
		return err
	default:
		return backoff.Permanent(err)
	}
}

// reaper returns an existing Reaper instance if it exists and is running, otherwise
// a new Reaper instance will be created with a sessionID to identify containers in
// the same test session/program. If connect is true, the reaper will be connected
// to the reaper container.
// Returns an error if config.RyukDisabled is true.
//
// Safe for concurrent calls.
func (r *reaperSpawner) reaper(ctx context.Context, sessionID string, provider ReaperProvider) (*Reaper, error) {
	if config.Read().RyukDisabled {
		return nil, errReaperDisabled
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	return backoff.RetryWithData(
		r.retryLocked(ctx, sessionID, provider),
		backoff.WithContext(r.backoff(), ctx),
	)
}

// retryLocked returns a function that can be used to create or reuse a reaper container.
// It must be called with the lock held.
func (r *reaperSpawner) retryLocked(ctx context.Context, sessionID string, provider ReaperProvider) func() (*Reaper, error) {
	return func() (*Reaper, error) {
		reaper, err := r.reuseOrCreate(ctx, sessionID, provider)
		if err != nil {
			return nil, r.retryError(errors.Join(err, r.cleanupLocked()))
		}

		r.instance = reaper

		return reaper, nil
	}
}

// reuseOrCreate returns an existing Reaper instance if it exists, otherwise a new Reaper instance.
func (r *reaperSpawner) reuseOrCreate(ctx context.Context, sessionID string, provider ReaperProvider) (*Reaper, error) {
	if r.instance != nil {
		// We already have an associated reaper, check it's still running and
		// we can still connect. A new reaper has just acknowledged a
		// connection, so needs neither.
		if err := r.isRunning(ctx, r.instance.container); err != nil {
			return nil, err
		}

		termSignal, err := r.instance.connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("connect: %w", err)
		}

		r.instance.setOrSignal(termSignal)

		return r.instance, nil
	}

	// Look for an existing reaper created in the same test session but in a
	// different test process execution e.g. when running tests in parallel.
	container, err := r.lookupContainer(ctx, sessionID, provider)
	if err != nil {
		if !errors.Is(err, errReaperNotFound) {
			return nil, fmt.Errorf("look up container: %w", err)
		}

		// The reaper container was not found, continue to create a new one.
		reaper, err := r.newReaper(ctx, sessionID, provider)
		if err != nil {
			return nil, fmt.Errorf("new reaper: %w", err)
		}

		return reaper, nil
	}

	// A reaper container exists re-use it.
	reaper, err := r.fromContainer(ctx, sessionID, provider, container)
	if err != nil {
		return nil, fmt.Errorf("from container %q: %w", container.ID[:8], err)
	}

	return reaper, nil
}

// fromContainer constructs a Reaper from an already running reaper DockerContainer.
func (r *reaperSpawner) fromContainer(ctx context.Context, sessionID string, provider ReaperProvider, dockerContainer *DockerContainer) (*Reaper, error) {
	log.Printf("⏳ Waiting for Reaper %q to be ready", dockerContainer.ID[:8])

	// Waited on first, as a reaper with no clients is counting down its
	// reconnection timeout. One that has stopped fails the wait straight away.
	conn := &reaperConn{}
	defer conn.Close()

	if err := r.waitStrategy(sessionID, conn).WaitUntilReady(ctx, dockerContainer); err != nil {
		// A reaper that stopped while waiting is also recreated by the retry.
		if stateErr := r.isRunning(ctx, dockerContainer); stateErr != nil {
			err = errors.Join(err, stateErr)
		} else if errors.Is(err, errReaperPortNotPublished) {
			// Docker never published its port, see docker/desktop-feedback#338,
			// so nothing can be using it.
			err = errors.Join(err, r.terminate(ctx, dockerContainer, StopTimeout(0)))
		}

		return nil, fmt.Errorf("wait for reaper %s: %w", dockerContainer.ID[:8], err)
	}

	log.Printf("🔥 Reaper obtained from Docker for this test session %s", dockerContainer.ID[:8])

	reaper := &Reaper{
		Provider:  provider,
		SessionID: sessionID,
		container: dockerContainer,
	}
	if err := reaper.keep(conn); err != nil {
		return nil, err
	}

	return reaper, nil
}

// defaultWaitStrategy returns the wait strategy that waits for a reaper container
// to acknowledge sessionID's filters, giving up after readyTimeout. The
// connection it was acknowledged on is kept in conn.
func (r *reaperSpawner) defaultWaitStrategy(sessionID string, conn *reaperConn) wait.Strategy {
	return wait.ForAll(&ryukHandshake{
		filters:        reaperFilters(sessionID),
		pollInterval:   100 * time.Millisecond,
		attemptTimeout: time.Second,
		conn:           conn,
	}).WithDeadline(r.readyTimeout)
}

// ryukHandshake is a wait strategy that waits for a reaper to acknowledge a
// session's filters, which only ryuk itself can do. Checking the port alone
// isn't enough, as the docker-proxy accepts connections on the published port
// whether or not ryuk is listening.
type ryukHandshake struct {
	filters        []byte
	pollInterval   time.Duration
	attemptTimeout time.Duration

	// conn keeps the connection the filters were acknowledged on, as the
	// reaper starts its reconnection timeout when it's closed.
	conn *reaperConn
}

// WaitUntilReady implements [wait.Strategy].
func (s *ryukHandshake) WaitUntilReady(ctx context.Context, target wait.StrategyTarget) error {
	for {
		// A reaper that has stopped, or whose port isn't published, will
		// never acknowledge, so is recreated rather than waited on. Docker
		// publishes ports before it reports a container as running.
		err := s.handshake(ctx, target)
		if err == nil || errdefs.IsNotFound(err) || errors.Is(err, errReaperPortNotPublished) {
			return err
		}

		select {
		case <-ctx.Done():
			return errors.Join(err, ctx.Err())
		case <-time.After(s.pollInterval):
		}
	}
}

// handshake sends the filters to target's reaper and reads its ACK. It
// returns a not found error if the reaper has stopped, or
// errReaperPortNotPublished if its port isn't published.
func (s *ryukHandshake) handshake(ctx context.Context, target wait.StrategyTarget) error {
	// Everything needed comes from one inspect, as the daemon may be slow.
	inspect, err := target.Inspect(ctx)
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}

	if !inspect.State.Running {
		return errdefs.ErrNotFound.WithMessage("container state: " + string(inspect.State.Status))
	}

	hostPort, err := reaperHostPort(inspect)
	if err != nil {
		return err
	}

	host, err := target.Host(ctx)
	if err != nil {
		return fmt.Errorf("host: %w", err)
	}

	// Ryuk acknowledges straight away, so give up on a connection that has
	// no response, e.g. one the docker-proxy holds open, and try again.
	ctx, cancel := context.WithTimeout(ctx, s.attemptTimeout)
	defer cancel()

	endpoint := net.JoinHostPort(host, hostPort)
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", endpoint)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	if err := acknowledge(ctx, conn, s.filters); err != nil {
		_ = conn.Close()
		return err
	}

	s.conn.set(conn, endpoint)

	return nil
}

// acknowledge sends filters on conn and reads the reaper's ACK, giving up
// when ctx is done.
func acknowledge(ctx context.Context, conn net.Conn, filters []byte) error {
	deadline, _ := ctx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set deadline: %w", err)
	}

	if _, err := conn.Write(filters); err != nil {
		return fmt.Errorf("write filters: %w", err)
	}

	buf := make([]byte, len(reaperAck))
	if _, err := io.ReadFull(conn, buf); err != nil {
		return fmt.Errorf("read ack: %w", err)
	}

	if !bytes.Equal(reaperAck, buf) {
		return fmt.Errorf("unexpected reaper response: %q", buf)
	}

	// The connection is kept, so mustn't time out.
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return fmt.Errorf("clear deadline: %w", err)
	}

	return nil
}

// String returns a human-readable description of the wait strategy.
func (s *ryukHandshake) String() string {
	return "reaper to acknowledge the session's filters"
}

// reaperConn holds a connection a reaper has acknowledged a session's filters
// on, and the endpoint it was made to, until it is taken or closed. It is safe
// for concurrent use.
type reaperConn struct {
	mtx      sync.Mutex
	conn     net.Conn
	endpoint string
}

// set makes conn, made to endpoint, the held connection, closing any it
// replaces.
func (c *reaperConn) set(conn net.Conn, endpoint string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.conn != nil {
		_ = c.conn.Close()
	}

	c.conn = conn
	c.endpoint = endpoint
}

// take returns the held connection and its endpoint, or nil if there isn't
// one. The caller is then responsible for closing the connection.
func (c *reaperConn) take() (net.Conn, string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	conn, endpoint := c.conn, c.endpoint
	c.conn, c.endpoint = nil, ""

	return conn, endpoint
}

// Close closes the held connection, if there is one.
func (c *reaperConn) Close() error {
	if conn, _ := c.take(); conn != nil {
		return conn.Close()
	}

	return nil
}

// reaperHostPort returns the host port the reaper container in inspect is
// published on, or errReaperPortNotPublished if it isn't published.
func reaperHostPort(inspect *container.InspectResponse) (string, error) {
	// The ryuk image exposes no ports, so the only port is the one the
	// reaper was created with, whichever process created it.
	var port network.Port
	for p := range inspect.NetworkSettings.Ports {
		if port.IsZero() || p.Num() < port.Num() {
			port = p
		}
	}

	for _, binding := range inspect.NetworkSettings.Ports[port] {
		if binding.HostPort != "" {
			return binding.HostPort, nil
		}
	}

	return "", errReaperPortNotPublished
}

// runReaper runs the reaper container described by req, logging to logger
// and giving up after timeout. conn is closed if the run fails or completes
// after its caller has returned. If ctx is done before the run completes, it
// returns ctx's error, and the run completes without its caller.
func (r *reaperSpawner) runReaper(ctx context.Context, provider ReaperProvider, req ContainerRequest, conn *reaperConn, logger *detachableLogger, timeout time.Duration) (Container, error) {
	type result struct {
		ctr Container
		err error
	}

	// Not cancelled with ctx, as abandoning a create the daemon has completed
	// leaves a reaper that never starts, blocking its name for the session.
	results := make(chan result)
	abandoned := make(chan struct{})
	go func() {
		runCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
		defer cancel()

		var res result
		c, err := provider.RunContainer(runCtx, req)
		if err != nil {
			// Remove whatever was created here, as the caller may have gone.
			// One whose port isn't published can't have been connected to, so
			// isn't waited on to stop.
			var opts []TerminateOption
			if errors.Is(err, errReaperPortNotPublished) {
				opts = append(opts, StopTimeout(0))
			}

			res.err = errors.Join(err, conn.Close(), r.terminate(ctx, c, opts...))
		} else {
			res.ctr = c
		}

		select {
		case results <- res:
		case <-abandoned:
			// Nothing will use the reaper, so let it exit.
			_ = conn.Close()

			// Report a failure nobody is waiting for, as it may have left a
			// container behind.
			if res.err != nil {
				id := "(none created)"
				if !isNil(c) {
					id = c.GetContainerID()
				}

				logger.Printf("🚫 Reaper container %s failed to run after its caller returned: %v", id, res.err)
			}
		}
	}()

	select {
	case res := <-results:
		return res.ctr, res.err
	case <-ctx.Done():
		logger.detach()
		close(abandoned)
		return nil, context.Cause(ctx)
	}
}

// terminate terminates ctr, if it isn't nil, with opts, giving up after the
// cleanup timeout.
func (r *reaperSpawner) terminate(ctx context.Context, ctr Container, opts ...TerminateOption) error {
	if isNil(ctr) {
		return nil
	}

	// Removed even if ctx is done, as its caller may have gone.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.cleanupTimeout)
	defer cancel()

	if err := ctr.Terminate(ctx, opts...); !isCleanupSafe(err) {
		return fmt.Errorf("terminate: %w", err)
	}

	return nil
}

// detachableLogger logs to its logger until it is detached, then to
// background.
type detachableLogger struct {
	mtx        sync.Mutex
	logger     log.Logger
	background log.Logger
}

// Printf implements log.Logger.
func (l *detachableLogger) Printf(format string, v ...any) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	l.logger.Printf(format, v...)
}

// detach makes the logger log to its background logger. When it returns, the
// previous logger is no longer in use.
func (l *detachableLogger) detach() {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	l.logger = l.background
}

// pullReaperImage pulls the image of req, a reaper request, if it isn't
// present, under the name CreateContainer would use for it.
func pullReaperImage(ctx context.Context, p *DockerProvider, req ContainerRequest) error {
	image, err := p.substituteImage(req.Image, p.imageSubstitutors(req))
	if err != nil {
		return err
	}

	if _, err := p.client.ImageInspect(ctx, image); err == nil {
		return nil
	} else if !errdefs.IsNotFound(err) {
		return fmt.Errorf("inspect image: %w", err)
	}

	return p.attemptToPullImage(ctx, image, client.ImagePullOptions{})
}

// newReaper creates a connected Reaper with a sessionID to identify containers
// and a provider to use.
func (r *reaperSpawner) newReaper(ctx context.Context, sessionID string, provider ReaperProvider) (reaper *Reaper, err error) {
	dockerHostMount := core.MustExtractDockerSocket(ctx)

	port := r.port()
	tcConfig := provider.Config().Config
	conn := &reaperConn{}
	defer conn.Close()

	req := ContainerRequest{
		Image:        config.ReaperDefaultImage,
		ExposedPorts: []string{port.String()},
		Labels:       core.DefaultLabels(sessionID),
		WaitingFor:   r.waitStrategy(sessionID, conn),
		Name:         reaperContainerNameFromSessionID(sessionID),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.AutoRemove = true
			hc.Binds = []string{dockerHostMount + ":/var/run/docker.sock"}
			hc.NetworkMode = Bridge
			hc.Privileged = tcConfig.RyukPrivileged
		},
		// Ryuk listens on RYUK_PORT, so it must match the port exposed.
		Env: map[string]string{
			"RYUK_PORT": port.Port(),
		},
	}
	if to := tcConfig.RyukConnectionTimeout; to > time.Duration(0) {
		req.Env["RYUK_CONNECTION_TIMEOUT"] = to.String()
	}
	if to := tcConfig.RyukReconnectionTimeout; to > time.Duration(0) {
		req.Env["RYUK_RECONNECTION_TIMEOUT"] = to.String()
	}
	if tcConfig.RyukVerbose {
		req.Env["RYUK_VERBOSE"] = "true"
	}

	// Setup reaper-specific labels for the reaper container.
	req.Labels[core.LabelReaper] = "true"
	req.Labels[core.LabelRyuk] = "true"
	delete(req.Labels, core.LabelReap)

	// The run can outlive its caller, whose logger may be a test logger, so
	// the run stops using it if the caller returns first.
	logger := &detachableLogger{background: r.detachedLogger}
	runProvider := provider

	// Attach reaper container to a requested network if it is specified
	if p, ok := provider.(*DockerProvider); ok {
		defaultNetwork, err := p.ensureDefaultNetwork(ctx)
		if err != nil {
			return nil, fmt.Errorf("ensure default network: %w", err)
		}

		req.Networks = append(req.Networks, defaultNetwork)

		// Unlike the run, the pull can't leave a container behind, so it uses
		// ctx rather than the run's timeout.
		if err := pullReaperImage(ctx, p, req); err != nil {
			return nil, fmt.Errorf("pull image: %w", err)
		}

		// A provider's logger is fixed when it's created, so the run gets a
		// provider of its own that logs to logger.
		logger.logger = p.Logger
		dp, err := NewDockerProvider(WithLogger(logger), WithDefaultBridgeNetwork(p.defaultBridgeNetworkName))
		if err != nil {
			return nil, fmt.Errorf("new provider: %w", err)
		}

		// Run on the caller's daemon, which may have been set with SetClient,
		// reached at the same host and on the same default network as its
		// other containers.
		dp.SetClient(p.client)
		if dp.hostCache, err = p.DaemonHost(ctx); err != nil {
			return nil, fmt.Errorf("daemon host: %w", err)
		}
		dp.defaultNetwork = defaultNetwork
		runProvider = dp
	}

	c, err := r.runReaper(ctx, runProvider, req, conn, logger, r.readyTimeout+reaperStartTimeout)
	defer func() {
		if err != nil {
			err = errors.Join(err, TerminateContainer(c))
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("run container: %w", err)
	}

	reaper = &Reaper{
		Provider:  provider,
		SessionID: sessionID,
		container: c,
	}
	if err = reaper.keep(conn); err != nil {
		return nil, err
	}

	return reaper, nil
}

// Reaper is used to start a sidecar container that cleans up resources
type Reaper struct {
	Provider   ReaperProvider
	SessionID  string
	Endpoint   string
	container  Container
	mtx        sync.Mutex // Protects termSignal.
	termSignal chan bool
}

// Connect connects to the reaper container and waits for it to acknowledge
// the labels sent to it, so that it cleans up the containers with them.
//
// It returns a channel that can be closed to terminate the connection.
// Returns an error if config.RyukDisabled is true, or if the reaper doesn't
// acknowledge the labels.
func (r *Reaper) Connect() (chan bool, error) {
	if config.Read().RyukDisabled {
		return nil, errReaperDisabled
	}

	if termSignal := r.useTermSignal(); termSignal != nil {
		return termSignal, nil
	}

	return r.connect(context.Background())
}

// close signals the connection to close if needed.
// Safe for concurrent calls.
func (r *Reaper) close() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.termSignal != nil {
		r.termSignal <- true
		r.termSignal = nil
	}
}

// setOrSignal sets the reapers termSignal field if nil
// otherwise consumes by sending true to it.
// Safe for concurrent calls.
func (r *Reaper) setOrSignal(termSignal chan bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.termSignal != nil {
		// Already have an existing connection, close the new one.
		termSignal <- true
		return
	}

	// First or new unused termSignal, assign for caller to reuse.
	r.termSignal = termSignal
}

// useTermSignal if termSignal is not nil returns it
// and sets it to nil, otherwise returns nil.
//
// Safe for concurrent calls.
func (r *Reaper) useTermSignal() chan bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.termSignal == nil {
		return nil
	}

	// Use existing connection.
	term := r.termSignal
	r.termSignal = nil

	return term
}

// connect connects to the reaper container and waits for it to acknowledge
// the labels sent to it, so that it cleans up the containers with them,
// giving up after the connect timeout.
//
// It returns a channel that can be sent true to terminate the connection.
func (r *Reaper) connect(ctx context.Context) (chan bool, error) {
	ctx, cancel := context.WithTimeout(ctx, reaperConnectTimeout)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", r.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("dial reaper %s: %w", r.Endpoint, err)
	}

	if err := acknowledge(ctx, conn, reaperFilters(r.SessionID)); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("handshake: %w", err)
	}

	return hold(conn), nil
}

// keep makes the connection held by conn, on which the reaper acknowledged
// r's filters, the connection Connect returns, and its endpoint r's. It
// returns an error if conn holds no connection.
func (r *Reaper) keep(conn *reaperConn) error {
	c, endpoint := conn.take()
	if c == nil {
		return errors.New("reaper acknowledged no connection")
	}

	r.Endpoint = endpoint
	r.setOrSignal(hold(c))

	return nil
}

// hold returns a channel that closes conn when it is sent to or closed.
func hold(conn net.Conn) chan bool {
	termSignal := make(chan bool)
	go func() {
		defer conn.Close()
		<-termSignal
	}()

	return termSignal
}

// reaperFilters returns the line sent to a reaper to register the filters for
// sessionID's resources. It is the same each time, so a reaper sent it more
// than once only registers the filters once.
func reaperFilters(sessionID string) []byte {
	labels := core.DefaultLabels(sessionID)
	labelFilters := make([]string, 0, len(labels))
	for l, v := range labels {
		labelFilters = append(labelFilters, fmt.Sprintf("label=%s=%s", l, v))
	}
	slices.Sort(labelFilters)

	return []byte(strings.Join(labelFilters, "&") + "\n")
}

// Deprecated: internally replaced by core.DefaultLabels(sessionID)
//
// Labels returns the container labels to use so that this Reaper cleans them up
func (r *Reaper) Labels() map[string]string {
	return GenericLabels()
}

// isReaperImage returns true if the image name is the reaper image.
func isReaperImage(name string) bool {
	return strings.HasSuffix(name, config.ReaperDefaultImage)
}
