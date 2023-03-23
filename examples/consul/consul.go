package consul

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// consulContainer represents the consul container type used in the module
type consulContainer struct {
	testcontainers.Container
	endpoint string
}

// startContainer creates an instance of the consul container type
func startContainer(ctx context.Context) (*consulContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "hashicorp/consul:latest",
		ExposedPorts: []string{"8500/tcp", "8600/udp"},
		Name:         "badger",
		Cmd:          []string{"agent", "-server", "-ui", "-node=server-1", "-bootstrap-expect=1", "-client=0.0.0.0"},
		WaitingFor:   wait.ForListeningPort("8500/tcp"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	mappedPort, err := container.MappedPort(ctx, "8500")
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	return &consulContainer{Container: container, endpoint: fmt.Sprintf("%s:%s", host, mappedPort.Port())}, nil
}
