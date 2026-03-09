package gcloud

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Deprecated: use [bigtable.Run] instead
// RunBigTableContainer creates an instance of the GCloud container type for BigTable.
func RunBigTableContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	return RunBigQuery(ctx, "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators", opts...)
}

// Deprecated: use [bigtable.Run] instead
// RunBigTable creates an instance of the GCloud container type for BigTable.
func RunBigTable(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("9000/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("9000/tcp"),
			wait.ForLog("running"),
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
		"gcloud beta emulators bigtable start --host-port 0.0.0.0:9000 --project="+settings.ProjectID,
	))

	return newGCloudContainer(ctx, img, 9000, settings, "", moduleOpts...)
}
