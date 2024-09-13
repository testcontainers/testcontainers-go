package vearch

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Container represents the Vearch container type used in the module
type Container struct {
	*testcontainers.DockerContainer
}

// Run creates an instance of the Vearch container type
func Run(ctx context.Context, img string, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	req := testcontainers.Request{
		Image:        img,
		ExposedPorts: []string{"8817/tcp", "9001/tcp"},
		Cmd:          []string{"-conf=/vearch/config.toml", "all"},
		Privileged:   true,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      "config.toml",
				ContainerFilePath: "/vearch/config.toml",
				FileMode:          0o666,
			},
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("8817/tcp").WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("9001/tcp").WithStartupTimeout(5*time.Second),
		),
		Started: true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	ctr, err := testcontainers.Run(ctx, req)
	var c *Container
	if ctr != nil {
		c = &Container{DockerContainer: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// RESTEndpoint returns the REST endpoint of the Vearch container
func (c *Container) RESTEndpoint(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "8817/tcp")
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host")
	}

	return fmt.Sprintf("http://%s:%s", host, containerPort.Port()), nil
}
