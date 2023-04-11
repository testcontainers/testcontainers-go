package nats

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultNATSImage = "nats:latest"
)

type natsContainer struct {
	testcontainers.Container
}

func (c *natsContainer) ConnectionString(ctx context.Context) (string, error) {
	port, err := c.MappedPort(ctx, "4222/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("nats://%s:%s", host, port.Port()), nil
}

func startContainer(ctx context.Context) (*natsContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultNATSImage,
		ExposedPorts: []string{"4222/tcp", "6222/tcp", "8222/tcp"},
		Cmd:          []string{"-DV", "-js"},
		WaitingFor:   wait.ForLog("Listening for client connections on 0.0.0.0:4222"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &natsContainer{Container: container}, nil
}
