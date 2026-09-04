package gcloud

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Deprecated: use [pubsub.Run] instead
// RunPubsubContainer creates an instance of the GCloud container type for Pubsub.
func RunPubsubContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	return RunPubsub(ctx, "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators", opts...)
}

// Deprecated: use [pubsub.Run] instead
// RunPubsub creates an instance of the GCloud container type for Pubsub.
func RunPubsub(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("8085/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("8085/tcp"),
			wait.ForLog("started"),
		),
	}

	moduleOpts = append(moduleOpts, opts...)

	settings, err := applyOptions(opts)
	if err != nil {
		return nil, err
	}

	moduleOpts = append(moduleOpts, testcontainers.WithCmd(
		"/bin/sh",
		"-c",
		"gcloud beta emulators pubsub start --host-port 0.0.0.0:8085 --project="+settings.ProjectID,
	))

	return newGCloudContainer(ctx, img, 8085, settings, "", moduleOpts...)
}
