// Package fakegcsserver provides a Testcontainers module for the
// Fake GCS Server (https://github.com/fsouza/fake-gcs-server), a
// Google Cloud Storage API emulator for local development and testing.
// It exposes the GCS JSON API on port 4443 with an in-memory storage backend.
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
// where scheme matches the value passed to [WithScheme] (default "http").
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

	// Build the HTTP wait strategy; enable TLS for "https" because port 4443
	// performs a TLS handshake in that mode.
	waitStrategy := wait.ForHTTP("/storage/v1/b").
		WithPort(defaultPort).
		WithStatusCodeMatcher(func(status int) bool {
			return status >= 200 && status < 500
		})
	if settings.Scheme == "https" {
		waitStrategy = waitStrategy.WithTLS(true).WithAllowInsecure(true)
	}

	// The image ENTRYPOINT is ["/bin/fake-gcs-server", "-data", "/data"].
	// WithCmd sets Docker CMD, which is appended after the entrypoint, so only
	// flags should be passed here — not the binary path.
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithCmd(
			"-scheme", settings.Scheme,
			"-port", "4443",
			"-backend", "memory",
		),
		testcontainers.WithWaitStrategy(waitStrategy),
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
