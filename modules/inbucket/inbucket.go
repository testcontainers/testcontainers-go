package inbucket

import (
	"context"
	"fmt"
	"net"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Container represents the Inbucket container type used in the module
type Container struct {
	*testcontainers.DockerContainer
}

// SmtpConnection returns the connection string for the smtp server, using the default
// 2500 port, and obtaining the host and exposed port from the container.
func (c *Container) SmtpConnection(ctx context.Context) (string, error) {
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
func (c *Container) WebInterface(ctx context.Context) (string, error) {
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

// Run creates an instance of the Inbucket container type
func Run(ctx context.Context, img string, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	req := testcontainers.Request{
		Image:        img,
		ExposedPorts: []string{"2500/tcp", "9000/tcp", "1100/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("2500/tcp"),
			wait.ForListeningPort("9000/tcp"),
			wait.ForListeningPort("1100/tcp"),
		),
		Started: true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	ctr, err := testcontainers.Run(ctx, req)
	if err != nil {
		return nil, err
	}

	return &Container{DockerContainer: ctr}, nil
}
