package gcloud

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Deprecated: use [firestore.Run] instead
// RunFirestoreContainer creates an instance of the GCloud container type for Firestore.
func RunFirestoreContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	return RunFirestore(ctx, "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators", opts...)
}

// Deprecated: use [firestore.Run] instead
// RunFirestore creates an instance of the GCloud container type for Firestore.
func RunFirestore(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort("8080/tcp"),
			wait.ForLog("running"),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	settings, err := applyOptions(opts)
	if err != nil {
		return nil, err
	}

	moduleOpts = append(moduleOpts, testcontainers.WithCmd(
		"/bin/sh",
		"-c",
		"gcloud beta emulators firestore start --host-port 0.0.0.0:8080 --project="+settings.ProjectID,
	))

	return newGCloudContainer(ctx, img, 8080, settings, "", moduleOpts...)
}
