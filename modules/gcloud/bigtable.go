package gcloud

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RunBigTableContainer creates an instance of the GCloud container type for BigTable
func RunBigTableContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
			ExposedPorts: []string{"9000/tcp"},
			WaitingFor:   wait.ForLog("running"),
		},
		Started: true,
	}

	settings := applyOptions(req, opts)

	req.Cmd = []string{
		"/bin/sh",
		"-c",
		"gcloud beta emulators bigtable start --host-port 0.0.0.0:9000 " + fmt.Sprintf("--project=%s", settings.ProjectID),
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	return newGCloudContainer(ctx, 9000, container, settings)
}
