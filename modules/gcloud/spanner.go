package gcloud

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Deprecated: use [spanner.Run] instead
// RunSpannerContainer creates an instance of the GCloud container type for Spanner.
func RunSpannerContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	return RunSpanner(ctx, "gcr.io/cloud-spanner-emulator/emulator:1.4.0", opts...)
}

// Deprecated: use [spanner.Run] instead
// RunSpanner creates an instance of the GCloud container type for Spanner.
func RunSpanner(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("9010/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("9010/tcp"),
			wait.ForLog("Cloud Spanner emulator running"),
		),
	}

	moduleOpts = append(moduleOpts, opts...)

	settings, err := applyOptions(opts)
	if err != nil {
		return nil, err
	}

	return newGCloudContainer(ctx, img, 9010, settings, "", moduleOpts...)
}
