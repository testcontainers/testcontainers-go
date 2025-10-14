package inbucket

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// InbucketContainer represents the Inbucket container type used in the module
type InbucketContainer struct {
	testcontainers.Container
}

// SmtpConnection returns the connection string for the smtp server, using the default
// 2500 port, and obtaining the host and exposed port from the container.
//
//nolint:revive,staticcheck //FIXME
func (c *InbucketContainer) SmtpConnection(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "2500/tcp", "")
}

// WebInterface returns the connection string for the web interface server,
// using the default 9000 port, and obtaining the host and exposed port from
// the container.
func (c *InbucketContainer) WebInterface(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "9000/tcp", "http")
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Inbucket container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*InbucketContainer, error) {
	return Run(ctx, "inbucket/inbucket:sha-2d409bb", opts...)
}

// Run creates an instance of the Inbucket container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*InbucketContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("2500/tcp", "9000/tcp", "1100/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("2500/tcp"),
			wait.ForListeningPort("9000/tcp"),
			wait.ForListeningPort("1100/tcp"),
		),
	}
	moduleOpts = append(moduleOpts, opts...)

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *InbucketContainer
	if container != nil {
		c = &InbucketContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("run inbucket: %w", err)
	}

	return c, nil
}
