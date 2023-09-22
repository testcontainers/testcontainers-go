package gcloud

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// DatastoreContainer represents the GCloud container type used in the module for Datastore
type DatastoreContainer struct {
	testcontainers.Container
	URI string
}

func (c *DatastoreContainer) uri(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, "8081")
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	return uri, nil
}

// RunDatastoreContainer creates an instance of the GCloud container type for Datastore
func RunDatastoreContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DatastoreContainer, error) {
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

	uri, err := (&DatastoreContainer{Container: container}).uri(ctx)
	if err != nil {
		return nil, err
	}

	return &DatastoreContainer{
		Container: container,
		URI:       uri,
	}, nil
}
