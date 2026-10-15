// Package sbx registers a Docker Sandbox runtime for Testcontainers.
// Importing the package does not start a sandbox; initialization is lazy.
package sbx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

// Provider owns one sandbox and its local Docker proxy.
type Provider struct {
	mu           sync.Mutex
	started      bool
	ready        chan struct{}
	config       testcontainers.RuntimeConfig
	err          error
	run          commandRunner
	name         string
	kitDir       string
	proxy        proxyServer
	proxyFactory func(context.Context, sandboxRuntime) (proxyServer, error)
}

var _ testcontainers.RuntimeProvider = (*Provider)(nil)

var defaultProvider Provider

func init() {
	if err := testcontainers.RegisterProvider("sbx", func() testcontainers.RuntimeProvider { return &defaultProvider }); err != nil {
		panic(err)
	}
}

// Initialize starts a unique sandbox once. Its result, including errors, is
// shared by all callers of this Provider.
func (p *Provider) Initialize(ctx context.Context) (testcontainers.RuntimeConfig, error) {
	if err := ctx.Err(); err != nil {
		return testcontainers.RuntimeConfig{}, err
	}
	p.mu.Lock()
	if !p.started {
		p.started = true
		p.ready = make(chan struct{})
		go p.initialize(context.WithoutCancel(ctx))
	}
	ready := p.ready
	p.mu.Unlock()
	select {
	case <-ctx.Done():
		return testcontainers.RuntimeConfig{}, ctx.Err()
	case <-ready:
		p.mu.Lock()
		defer p.mu.Unlock()
		return p.config, p.err
	}
}

func (p *Provider) initialize(parent context.Context) {
	var config testcontainers.RuntimeConfig
	var err error
	defer func() {
		p.mu.Lock()
		p.config, p.err = config, err
		close(p.ready)
		p.mu.Unlock()
	}()
	cfg, configErr := readConfiguration()
	if configErr != nil {
		err = configErr
		return
	}
	ctx, cancel := context.WithTimeout(parent, cfg.timeout)
	defer cancel()
	run := newCommandRunner(cfg)
	p.mu.Lock()
	p.run = run
	p.mu.Unlock()

	name, nameErr := newSandboxName()
	if nameErr != nil {
		err = nameErr
		return
	}
	kitDir, kitErr := materializeKit()
	if kitErr != nil {
		err = kitErr
		return
	}
	p.mu.Lock()
	p.kitDir = kitDir
	p.mu.Unlock()

	// Probe and release an IPv4 loopback port immediately before create. This
	// is advisory: sbx's actual bind decides whether the port is still free.
	hostPort, portErr := candidateHostPort(ctx, "tcp", nil, nil)
	if portErr != nil {
		err = errors.Join(fmt.Errorf("choose sbx Docker API host port: %w", portErr), p.cleanupAfterFailure())
		return
	}
	p.mu.Lock()
	// Reserve the generated name before invoking create. A timeout can leave
	// a sandbox behind even when the CLI returns an error.
	p.name = name
	p.mu.Unlock()

	// create rejects existing names, unlike run, and exits after creating a
	// running sandbox. Its stdout is intentionally ignored.
	_, err = run(ctx, "create", "--name", name, "--kit", kitDir,
		"--publish", fmt.Sprintf("127.0.0.1:%d:2375/tcp", hostPort), baseImage, cfg.workspace)
	if err != nil {
		if isCreateCollision(err, name) {
			// create checks for an existing name before issuing its create call.
			// That sandbox is not ours, even with an unlikely UUID collision.
			p.mu.Lock()
			p.name = ""
			p.mu.Unlock()
		}
		err = errors.Join(err, p.cleanupAfterFailure())
		return
	}

	dockerURL, discoverErr := discoverDockerURL(ctx, run, name, hostPort)
	if discoverErr != nil {
		err = errors.Join(discoverErr, p.cleanupAfterFailure())
		return
	}
	if pingErr := waitForDocker(ctx, dockerURL); pingErr != nil {
		err = errors.Join(pingErr, p.cleanupAfterFailure())
		return
	}
	factory := p.proxyFactory
	if factory == nil {
		factory = newDockerProxy
	}
	proxy, proxyErr := factory(ctx, sandboxRuntime{name: name, run: run, dockerURL: dockerURL, workspace: cfg.workspace})
	if proxyErr != nil {
		err = errors.Join(fmt.Errorf("start sbx Docker proxy: %w", proxyErr), p.cleanupAfterFailure())
		return
	}
	p.mu.Lock()
	p.proxy = proxy
	p.mu.Unlock()
	config = testcontainers.RuntimeConfig{
		Name: "sbx", DockerHost: proxy.Host(), DockerSocketPath: "/var/run/docker.sock",
		PublishedHost: "127.0.0.1", DefaultBridgeNetwork: "bridge",
	}
}

func (p *Provider) cleanupAfterFailure() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return p.cleanupResources(ctx)
}

func isCreateCollision(err error, name string) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("sandbox '%s' already exists; connect to the existing sandbox", name))
}

func isMissingSandbox(err error, name string) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("sandbox '%s' not found", name))
}

// Cleanup removes the default provider's owned sandbox if it was initialized.
func Cleanup(ctx context.Context) error { return defaultProvider.Cleanup(ctx) }

// Cleanup is safe to retry if the sandbox removal fails.
func (p *Provider) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		return nil
	}
	ready := p.ready
	p.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ready:
		return p.cleanupResources(ctx)
	}
}

func (p *Provider) cleanupResources(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var errs []error
	if p.proxy != nil {
		if err := p.proxy.Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("close sbx Docker proxy: %w", err))
		} else {
			p.proxy = nil
		}
	}
	if p.name != "" {
		if _, err := p.run(ctx, "rm", "--force", p.name); err != nil && !isMissingSandbox(err, p.name) {
			errs = append(errs, fmt.Errorf("remove owned sbx sandbox %q: %w", p.name, err))
		} else {
			p.name = ""
		}
	}
	if p.name == "" && p.kitDir != "" {
		if err := os.RemoveAll(p.kitDir); err != nil {
			errs = append(errs, fmt.Errorf("remove sbx kit directory: %w", err))
		} else {
			p.kitDir = ""
		}
	}
	return errors.Join(errs...)
}

func newSandboxName() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", fmt.Errorf("generate sbx sandbox name: %w", err)
	}
	id[6] = id[6]&0x0f | 0x40
	id[8] = id[8]&0x3f | 0x80
	return fmt.Sprintf("tc-go-%s-%s-%s-%s-%s", hex.EncodeToString(id[:4]), hex.EncodeToString(id[4:6]), hex.EncodeToString(id[6:8]), hex.EncodeToString(id[8:10]), hex.EncodeToString(id[10:])), nil
}

type publishedPort struct {
	HostIP      string `json:"host_ip"`
	HostPort    int    `json:"host_port"`
	SandboxPort int    `json:"sandbox_port"`
	Protocol    string `json:"protocol"`
}

func discoverDockerURL(ctx context.Context, run commandRunner, name string, expectedPort int) (*url.URL, error) {
	if expectedPort < 1 || expectedPort > 65535 {
		return nil, fmt.Errorf("invalid expected sbx Docker host port %d", expectedPort)
	}
	endpoint := net.JoinHostPort("127.0.0.1", strconv.Itoa(expectedPort))
	output, err := run(ctx, "ports", name, "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("discover sbx Docker port: %w", err)
	}
	var ports []publishedPort
	if err := json.Unmarshal(output, &ports); err != nil {
		return nil, fmt.Errorf("decode sbx published ports: %w", err)
	}
	found := false
	for _, port := range ports {
		if port.HostIP != "127.0.0.1" || port.HostPort != expectedPort || (port.Protocol != "tcp" && port.Protocol != "tcp4") {
			continue
		}
		if port.SandboxPort != 2375 {
			return nil, fmt.Errorf("conflicting sbx Docker mapping for %s: sandbox port %d, want 2375/tcp", endpoint, port.SandboxPort)
		}
		// tcp and tcp4 observations of this same IPv4 socket are equivalent.
		found = true
	}
	if !found {
		return nil, fmt.Errorf("sbx did not publish Docker port 2375/tcp on %s", endpoint)
	}
	return &url.URL{Scheme: "http", Host: endpoint}, nil
}

func waitForDocker(ctx context.Context, endpoint *url.URL) error {
	transport := &http.Transport{Proxy: nil, DialContext: (&net.Dialer{Timeout: 2 * time.Second}).DialContext}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}
	url := endpoint.JoinPath("/_ping").String()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	var lastErr error
	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("make sbx Docker ping request: %w", err)
		}
		response, err := client.Do(request)
		if err == nil {
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1024))
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
			lastErr = fmt.Errorf("HTTP %d", response.StatusCode)
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for sbx Docker API at %s: %w (last response: %w)", endpoint.Host, ctx.Err(), lastErr)
		case <-ticker.C:
		}
	}
}
