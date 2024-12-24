package gcloud

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Deprecated: use RunFirestore instead
// RunFirestoreContainer creates an instance of the GCloud container type for Firestore.
func RunFirestoreContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	return RunFirestore(ctx, "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators", opts...)
}

// RunFirestore creates an instance of the GCloud container type for Firestore.
func RunFirestore(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        img,
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor:   wait.ForLog("running"),
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
		"gcloud beta emulators firestore start --host-port 0.0.0.0:8080 --project=" + settings.ProjectID,
	}

	return newGCloudContainer(ctx, req, 8080, settings, "")
}
