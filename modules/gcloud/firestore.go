package gcloud

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// FirestoreContainer represents the firestore container type used in the module
type FirestoreContainer struct {
	testcontainers.Container
	URI string
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
	req := testcontainers.ContainerRequest{
		Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("running"),
		Cmd: []string{
			"/bin/sh",
			"-c",
			"gcloud beta emulators firestore start --host-port 0.0.0.0:8080",
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	uri, err := (&FirestoreContainer{Container: container}).uri(ctx)
	if err != nil {
		return nil, err
	}

	return &FirestoreContainer{
		Container: container,
		URI:       uri,
	}, nil
}
