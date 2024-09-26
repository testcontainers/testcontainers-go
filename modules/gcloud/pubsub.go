package gcloud

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Deprecated: use RunPubsub instead
// RunPubsubContainer creates an instance of the GCloud container type for Pubsub.
func RunPubsubContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	return RunPubsub(ctx, "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators", opts...)
}

// RunPubsub creates an instance of the GCloud container type for Pubsub.
func RunPubsub(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        img,
			ExposedPorts: []string{"8085/tcp"},
			WaitingFor:   wait.ForLog("started"),
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
		"gcloud beta emulators pubsub start --host-port 0.0.0.0:8085 --project=" + settings.ProjectID,
	}

	return newGCloudContainer(ctx, req, 8085, settings, "")
}
