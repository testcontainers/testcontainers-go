package memcached

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort = "11211/tcp"
)

// Container represents the Memcached container type used in the module
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the Memcached container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(defaultPort)),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run memcached: %w", err)
	}

	return c, nil
}

// HostPort returns the host and port of the Memcached container
func (c *Container) HostPort(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, defaultPort, "")
}
