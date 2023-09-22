package gcloud

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// FirestoreContainer represents the GCloud container type used in the module for Firestore
type FirestoreContainer struct {
	testcontainers.Container
	Settings options
	URI      string
}

func (c *FirestoreContainer) uri(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, "8080")
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

// RunFirestoreContainer creates an instance of the GCloud container type for Firestore
func RunFirestoreContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*FirestoreContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor:   wait.ForLog("running"),
		},
		Started: true,
	}

	settings := applyOptions(req, opts)

	req.Cmd = []string{
		"/bin/sh",
		"-c",
		"gcloud beta emulators firestore start --host-port 0.0.0.0:8080 " + fmt.Sprintf("--project=%s", settings.ProjectID),
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	firestoreContainer := FirestoreContainer{
		Container: container,
		Settings:  settings,
	}

	uri, err := containerURI(ctx, &firestoreContainer)
	if err != nil {
		return nil, err
	}

	firestoreContainer.URI = uri

	return &firestoreContainer, nil
}
