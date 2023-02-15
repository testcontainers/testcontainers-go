package datastore

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// datastoreContainer represents the datastore container type used in the module
type datastoreContainer struct {
	testcontainers.Container
	URI string
}

// startContainer creates an instance of the datastore container type
func startContainer(ctx context.Context) (*datastoreContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		ExposedPorts: []string{"8081/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort("8081/tcp"),
		Cmd: []string{
			"/bin/sh",
			"-c",
			"gcloud beta emulators datastore start --project test-project --host-port 0.0.0.0:8081",
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "8081")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())

	return &datastoreContainer{Container: container, URI: uri}, nil
}
