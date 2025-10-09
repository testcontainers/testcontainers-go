package spanner

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// DefaultProjectID is the default project ID for the Spanner container.
	DefaultProjectID  = "test-project"
	defaultPortNumber = "9010"
	defaultPort       = defaultPortNumber + "/tcp"
)

// Container represents the Spanner container type used in the module
type Container struct {
	testcontainers.Container
	settings options
}

// ProjectID returns the project ID of the Spanner container.
func (c *Container) ProjectID() string {
	return c.settings.ProjectID
}

// URI returns the URI of the Spanner container.
func (c *Container) URI() string {
	return c.settings.URI
}

// Run creates an instance of the Spanner GCloud container type.
// The URI uses the empty string as the protocol.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort(defaultPort),
			wait.ForLog("Cloud Spanner emulator running"),
		)),
	}

	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, err
			}
		}
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, settings: settings}
	}
	if err != nil {
		return c, fmt.Errorf("run spanner: %w", err)
	}

	portEndpoint, err := c.PortEndpoint(ctx, defaultPort, "")
	if err != nil {
		return c, fmt.Errorf("port endpoint: %w", err)
	}

	c.settings.URI = portEndpoint

	return c, nil
}
