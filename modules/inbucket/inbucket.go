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

	return fmt.Sprintf("http://%s", net.JoinHostPort(host, containerPort.Port())), nil
}

// RunContainer creates an instance of the Inbucket container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*InbucketContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "inbucket/inbucket:sha-2d409bb",
		ExposedPorts: []string{"2500/tcp", "9000/tcp"},
		WaitingFor:   wait.ForLog("SMTP listening on tcp4"),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &InbucketContainer{Container: container}, nil
}
