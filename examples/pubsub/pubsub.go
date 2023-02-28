package pubsub

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// pubsubContainer represents the pubsub container type used in the module
type pubsubContainer struct {
	testcontainers.Container
	URI string
}

// startContainer creates an instance of the pubsub container type
func startContainer(ctx context.Context) (*pubsubContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		ExposedPorts: []string{"8085/tcp"},
		WaitingFor:   wait.ForLog("started"),
		Cmd: []string{
			"/bin/sh",
			"-c",
			"gcloud beta emulators pubsub start --host-port 0.0.0.0:8085",
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "8085")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())

	return &pubsubContainer{Container: container, URI: uri}, nil
}
