package vearch

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// VearchContainer represents the Vearch container type used in the module
type VearchContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Vearch container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*VearchContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "vearch/vearch:latest",
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
	if err != nil {
		return nil, err
	}

	return &VearchContainer{Container: container}, nil
}

// RESTEndpoint returns the REST endpoint of the Vearch container
func (c *VearchContainer) RESTEndpoint(ctx context.Context) (string, error) {
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
