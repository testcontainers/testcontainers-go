package gcloud

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Deprecated: use [datastore.Run] instead
// RunDatastoreContainer creates an instance of the GCloud container type for Datastore.
func RunDatastoreContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	return RunDatastore(ctx, "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators", opts...)
}

// Deprecated: use [datastore.Run] instead
// RunDatastore creates an instance of the GCloud container type for Datastore.
func RunDatastore(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        img,
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
		"gcloud beta emulators datastore start --host-port 0.0.0.0:8081 --project=" + settings.ProjectID,
	}

	return newGCloudContainer(ctx, req, 8081, settings, "")
}
