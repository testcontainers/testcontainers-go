package spanner

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/testcontainers/testcontainers-go"
)

// spannerContainer represents the spanner container type used in the module
type spannerContainer struct {
	testcontainers.Container
	GRPCEndpoint string
}

// startContainer creates an instance of the spanner container type
func startContainer(ctx context.Context) (*spannerContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "gcr.io/cloud-spanner-emulator/emulator:1.4.0",
		ExposedPorts: []string{"9010/tcp"},
		WaitingFor:   wait.ForLog("Cloud Spanner emulator running"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "9010")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	grpcEndpoint := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())

	return &spannerContainer{Container: container, GRPCEndpoint: grpcEndpoint}, nil
}
