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
	Settings options
	URI      string
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
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
			ExposedPorts: []string{"8081/tcp"},
			WaitingFor:   wait.ForHTTP("/").WithPort("8081/tcp"),
		},
		Started: true,
	}

	settings := applyOptions(req, opts)

	req.Cmd = []string{
		"/bin/sh",
		"-c",
		"gcloud beta emulators datastore start --host-port 0.0.0.0:8081 " + fmt.Sprintf("--project=%s", settings.ProjectID),
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	datastoreContainer := DatastoreContainer{
		Container: container,
		Settings:  settings,
	}

	uri, err := containerURI(ctx, &datastoreContainer)
	if err != nil {
		return nil, err
	}

	datastoreContainer.URI = uri

	return &datastoreContainer, nil
}
