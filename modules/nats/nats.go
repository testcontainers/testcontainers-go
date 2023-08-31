package nats

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NATSContainer represents the NATS container type used in the module
type NATSContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the NATS container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*NATSContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "nats:2.9",
		ExposedPorts: []string{"4222/tcp", "6222/tcp", "8222/tcp"},
		Cmd:          []string{"-DV", "-js"},
		WaitingFor:   wait.ForLog("Listening for client connections on 0.0.0.0:4222"),
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

	return &NATSContainer{Container: container}, nil
}

// ConnectionString returns a connection string for the NATS container
func (c *NATSContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	mappedPort, err := c.MappedPort(ctx, "4222/tcp")
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	uri := fmt.Sprintf("nats://%s:%s", hostIP, mappedPort.Port())
	return uri, nil
}
