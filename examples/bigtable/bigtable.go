package bigtable

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/testcontainers/testcontainers-go"
)

// bigtableContainer represents the bigtable container type used in the module
type bigtableContainer struct {
	testcontainers.Container
	URI string
}

// startContainer creates an instance of the bigtable container type
func startContainer(ctx context.Context) (*bigtableContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		ExposedPorts: []string{"9000/tcp"},
		WaitingFor:   wait.ForLog("running"),
		Cmd: []string{
			"/bin/sh",
			"-c",
			"gcloud beta emulators bigtable start --host-port 0.0.0.0:9000",
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "9000")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())

	return &bigtableContainer{Container: container, URI: uri}, nil
}
