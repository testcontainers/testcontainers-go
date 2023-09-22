package gcloud

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PubsubContainer represents the pubsub container type used in the module
type PubsubContainer struct {
	testcontainers.Container
	URI string
}

func (c *PubsubContainer) uri(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, "8085")
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

// RunPubsubContainer creates an instance of the GCloud container type for Pubsub
func RunPubsubContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*PubsubContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		ExposedPorts: []string{"8085/tcp"},
		WaitingFor:   wait.ForLog("started"),
		Cmd: []string{
			"/bin/sh",
			"-c",
			"gcloud beta emulators pubsub start --host-port 0.0.0.0:8085",
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	pubsubContainer := PubsubContainer{
		Container: container,
	}

	uri, err := containerURI(ctx, &pubsubContainer)
	if err != nil {
		return nil, err
	}

	pubsubContainer.URI = uri

	return &pubsubContainer, nil
}
