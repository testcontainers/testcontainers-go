package firestore

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/testcontainers/testcontainers-go"
)

// firestoreContainer represents the firestore container type used in the module
type firestoreContainer struct {
	testcontainers.Container
	URI string
}

// startContainer creates an instance of the firestore container type
func startContainer(ctx context.Context) (*firestoreContainer, error) {
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

	mappedPort, err := container.MappedPort(ctx, "8080")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())

	return &firestoreContainer{Container: container, URI: uri}, nil
}
