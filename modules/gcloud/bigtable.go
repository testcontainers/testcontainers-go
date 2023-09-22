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
	Settings options
	URI      string
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
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
			ExposedPorts: []string{"9000/tcp"},
			WaitingFor:   wait.ForLog("running"),
		},
		Started: true,
	}

	settings := applyOptions(req, opts)

	req.Cmd = []string{
		"/bin/sh",
		"-c",
		"gcloud beta emulators bigtable start --host-port 0.0.0.0:9000 " + fmt.Sprintf("--project=%s", settings.ProjectID),
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	bigTableContainer := BigTableContainer{
		Container: container,
		Settings:  settings,
	}

	uri, err := containerURI(ctx, &bigTableContainer)
	if err != nil {
		return nil, err
	}

	bigTableContainer.URI = uri

	return &bigTableContainer, nil
}
