package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/testcontainers/testcontainers-go/internal/config"
)

// RuntimeProvider prepares a Docker-compatible runtime before client discovery.
// Initialize must not call testcontainers client or provider constructors, which
// would recursively initialize the runtime. It is called at most once per process.
// Its context preserves the first caller's values but not cancellation or deadlines;
// implementations must bound startup with their own timeout when needed.
type RuntimeProvider interface {
	Initialize(context.Context) (RuntimeConfig, error)
}

// RuntimeConfig separates the client endpoint, daemon-side socket mount, and
// address at which the test process can reach published container ports.
// Empty endpoint fields retain the standard Docker discovery behavior.
type RuntimeConfig struct {
	Name                 string
	DockerHost           string
	DockerSocketPath     string
	PublishedHost        string
	DefaultBridgeNetwork string
}

// ProviderFactory constructs a runtime without initializing it during import.
type ProviderFactory func() RuntimeProvider

var runtimeProviders = newRuntimeRegistry()

type runtimeRegistry struct {
	mu        sync.Mutex
	factories map[string]ProviderFactory
	requested string
	started   bool
	name      string
	ready     chan struct{}
	config    RuntimeConfig
	err       error
}

func newRuntimeRegistry() *runtimeRegistry {
	return &runtimeRegistry{factories: make(map[string]ProviderFactory)}
}

func (r *runtimeRegistry) register(name string, factory ProviderFactory) error {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || name == "docker" || factory == nil {
		return fmt.Errorf("invalid provider registration %q", name)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		return errors.New("register providers before Docker discovery starts")
	}
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("provider %q is already registered", name)
	}
	r.factories[name] = factory
	return nil
}

func (r *runtimeRegistry) selectProvider(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		// Legacy Docker and Podman selections share the same discovered daemon.
		// Keep their per-request network behavior during the migration.
		if name == r.name || (isLegacyProvider(name) && isLegacyProvider(r.name)) {
			return nil
		}
		return fmt.Errorf("provider %q is already initialized; cannot switch to %q", r.name, name)
	}
	r.requested = name
	return nil
}

func (r *runtimeRegistry) resolve(ctx context.Context, name string) (RuntimeConfig, error) {
	if err := ctx.Err(); err != nil {
		return RuntimeConfig{}, err
	}
	r.mu.Lock()
	if !r.started {
		if r.requested != "" {
			name = r.requested
		}
		r.started = true
		r.name = name
		r.ready = make(chan struct{})
		factory := r.factories[name]
		// Startup is shared by the process. A caller may stop waiting without
		// canceling initialization for every other caller or starting it twice.
		go func() {
			cfg, err := initializeRuntime(context.WithoutCancel(ctx), name, factory)
			r.mu.Lock()
			r.config, r.err = cfg, err
			close(r.ready)
			r.mu.Unlock()
		}()
	}
	ready := r.ready
	r.mu.Unlock()

	select {
	case <-ctx.Done():
		return RuntimeConfig{}, ctx.Err()
	case <-ready:
		return r.config, r.err
	}
}

func initializeRuntime(ctx context.Context, name string, factory ProviderFactory) (RuntimeConfig, error) {
	cfg := RuntimeConfig{Name: name, DefaultBridgeNetwork: "bridge"}
	if factory == nil {
		switch name {
		case "docker":
			return cfg, nil
		case "podman":
			// Compatibility path for users who have not imported modules/podman.
			cfg.DefaultBridgeNetwork = "podman"
			return cfg, nil
		default:
			return RuntimeConfig{}, fmt.Errorf("provider %q is not registered; import its Go module before using it", name)
		}
	}
	p := factory()
	if p == nil {
		return RuntimeConfig{}, fmt.Errorf("provider %q factory returned nil", name)
	}
	cfg, err := p.Initialize(ctx)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("initialize provider %q: %w", name, err)
	}
	cfg.Name = name
	if cfg.DefaultBridgeNetwork == "" {
		cfg.DefaultBridgeNetwork = "bridge"
	}
	return cfg, nil
}

func isLegacyProvider(name string) bool {
	return name == "docker" || name == "podman"
}

// ConfiguredProviderName resolves a name without discovering a Docker endpoint.
func ConfiguredProviderName() string {
	if name := strings.ToLower(strings.TrimSpace(config.Read().Provider)); name != "" {
		return name
	}
	if strings.Contains(os.Getenv("DOCKER_HOST"), "podman.sock") {
		return "podman"
	}
	return "docker"
}

func RegisterProvider(name string, factory ProviderFactory) error {
	return runtimeProviders.register(name, factory)
}

func SelectRuntime(name string) error {
	return runtimeProviders.selectProvider(name)
}

func Runtime(ctx context.Context) (RuntimeConfig, error) {
	return runtimeProviders.resolve(ctx, ConfiguredProviderName())
}

func HasRegisteredProvider(name string) bool {
	runtimeProviders.mu.Lock()
	defer runtimeProviders.mu.Unlock()
	_, ok := runtimeProviders.factories[name]
	return ok
}
