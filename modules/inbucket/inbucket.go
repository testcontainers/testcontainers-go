package inbucket

import (
	"context"
	"fmt"
	"net"

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
	containerPort, err := c.MappedPort(ctx, "2500/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	return net.JoinHostPort(host, containerPort.Port()), nil
}

// WebInterface returns the connection string for the web interface server,
// using the default 9000 port, and obtaining the host and exposed port from
// the container.
func (c *InbucketContainer) WebInterface(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "9000/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	return "http://" + net.JoinHostPort(host, containerPort.Port()), nil
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Inbucket container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*InbucketContainer, error) {
	return Run(ctx, "inbucket/inbucket:sha-2d409bb", opts...)
}

// Run creates an instance of the Inbucket container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*InbucketContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"2500/tcp", "9000/tcp", "1100/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("2500/tcp"),
			wait.ForListeningPort("9000/tcp"),
			wait.ForListeningPort("1100/tcp"),
		),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *InbucketContainer
	if container != nil {
		c = &InbucketContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
