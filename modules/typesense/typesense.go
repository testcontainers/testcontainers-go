package typesense

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultAPIKey   = "test-api-key"
	defaultDataDir  = "/tmp"
	defaultHTTPPort = "8108/tcp"
	apiKeyEnvVar    = "TYPESENSE_API_KEY"
	dataDirEnvVar   = "TYPESENSE_DATA_DIR"
)

// Container represents the Typesense container type used in the module
type Container struct {
	testcontainers.Container
	apiKey string
}

// options holds configuration for the Typesense container.
type options struct {
	apiKey string
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is a function that configures the Typesense container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	return nil
}

// WithAPIKey sets the API key for the Typesense container.
// The API key is required for all Typesense API calls.
func WithAPIKey(key string) Option {
	return func(o *options) {
		o.apiKey = key
	}
}

// WithDataDir sets the data directory for the Typesense container.
func WithDataDir(dir string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		dataDirEnvVar: dir,
	})
}

// Run creates an instance of the Typesense container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// Gather module-specific options and apply defaults.
	settings := &options{
		apiKey: defaultAPIKey,
	}
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(settings)
		}
	}

	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultHTTPPort),
		testcontainers.WithEnv(map[string]string{
			dataDirEnvVar: defaultDataDir,
			apiKeyEnvVar:  settings.apiKey,
		}),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/health").
				WithPort(defaultHTTPPort).
				WithStatusCodeMatcher(func(status int) bool {
					return status == http.StatusOK
				}),
		),
	}

	ctr, err := testcontainers.Run(ctx, img, append(moduleOpts, opts...)...)
	var c *Container
	if ctr != nil {
		c = &Container{
			Container: ctr,
			apiKey:    settings.apiKey,
		}
	}

	if err != nil {
		return c, fmt.Errorf("run typesense: %w", err)
	}

	// Inspect the container to get the effective API key, which may have been
	// overridden by the caller via testcontainers.WithEnv after module defaults.
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect typesense: %w", err)
	}
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, apiKeyEnvVar+"="); ok {
			c.apiKey = v
		}
	}

	return c, nil
}

// Address retrieves the HTTP address of the Typesense container.
func (c *Container) Address(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, defaultHTTPPort, "http")
}

// APIKey returns the API key configured for the Typesense container.
// The API key is required for all Typesense API calls.
func (c *Container) APIKey() string {
	return c.apiKey
}
