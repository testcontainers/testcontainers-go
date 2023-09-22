package gcloud

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// BigTableContainer represents the GCloud container type used in the module for BigTable
type BigTableContainer struct {
	testcontainers.Container
	URI string
}

func (c *BigTableContainer) uri(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, "9000")
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

// RunBigTableContainer creates an instance of the GCloud container type for BigTable
func RunBigTableContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*BigTableContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		ExposedPorts: []string{"9000/tcp"},
		WaitingFor:   wait.ForLog("running"),
		Cmd: []string{
			"/bin/sh",
			"-c",
			"gcloud beta emulators bigtable start --host-port 0.0.0.0:9000",
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	uri, err := (&BigTableContainer{Container: container}).uri(ctx)
	if err != nil {
		return nil, err
	}

	return &BigTableContainer{
		Container: container,
		URI:       uri,
	}, nil
}
