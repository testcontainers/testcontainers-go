package fakegcsserver

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// defaultPort is the port on which the fake-gcs-server listens.
	defaultPort = "4443/tcp"
)

// Container represents the FakeGCSServer container type used in the module.
type Container struct {
	testcontainers.Container
	opts options
}

// StorageURL returns the GCS-compatible storage URL for the container.
// The URL is in the form: <scheme>://<host>:<port>/storage/v1
func (c *Container) StorageURL(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("get host: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", fmt.Errorf("get mapped port: %w", err)
	}

	return fmt.Sprintf("%s://%s:%s/storage/v1", c.opts.Scheme, host, port.Port()), nil
}

// Run creates an instance of the FakeGCSServer container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	settings := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&settings); err != nil {
				return nil, fmt.Errorf("fakegcsserver option: %w", err)
			}
		}
	}

	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithCmd(
			"/bin/fake-gcs-server",
			"-scheme", settings.Scheme,
			"-port", "4443",
			"-backend", "memory",
		),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/storage/v1/b").
				WithPort(defaultPort).
				WithStatusCodeMatcher(func(status int) bool {
					return status == 200
				}),
		),
	)
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, opts: settings}
	}

	if err != nil {
		return c, fmt.Errorf("run fakegcsserver: %w", err)
	}

	return c, nil
}
