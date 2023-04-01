package rabbitmq

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// rabbitmqContainer represents the mongodb container type used in the module
type rabbitmqContainer struct {
	testcontainers.Container
	endpoint string
}

func startContainer(ctx context.Context) (*rabbitmqContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForLog("started TCP listener on [::]:5672"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "5672")
	if err != nil {
		return nil, err
	}

	return &rabbitmqContainer{
		Container: container,
		endpoint:  fmt.Sprintf("%s:%s", host, mappedPort.Port()),
	}, nil
}
