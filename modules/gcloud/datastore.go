package gcloud

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RunDatastoreContainer creates an instance of the GCloud container type for Datastore
func RunDatastoreContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
			ExposedPorts: []string{"8081/tcp"},
			WaitingFor:   wait.ForHTTP("/").WithPort("8081/tcp"),
		},
		Started: true,
	}

	settings, err := applyOptions(&req, opts)
	if err != nil {
		return nil, err
	}

	req.Cmd = []string{
		"/bin/sh",
		"-c",
		"gcloud beta emulators datastore start --host-port 0.0.0.0:8081 " + fmt.Sprintf("--project=%s", settings.ProjectID),
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	return newGCloudContainer(ctx, 8081, container, settings)
}
