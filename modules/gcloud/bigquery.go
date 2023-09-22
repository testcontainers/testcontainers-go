package gcloud

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

// BigQueryContainer represents the GCloud container type used in the module for BigQuery
type BigQueryContainer struct {
	testcontainers.Container
	URI string
}

func (c *BigQueryContainer) uri(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, "9050")
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	uri := fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())
	return uri, nil
}

// RunBigQueryContainer creates an instance of the GCloud container type for BigQuery
func RunBigQueryContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*BigQueryContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/goccy/bigquery-emulator:0.4.3",
		ExposedPorts: []string{"9050/tcp", "9060/tcp"},
		Cmd:          []string{"--project", "test-project"},
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

	bigQueryContainer := BigQueryContainer{
		Container: container,
	}

	uri, err := containerURI(ctx, &bigQueryContainer)
	if err != nil {
		return nil, err
	}

	bigQueryContainer.URI = uri

	return &bigQueryContainer, nil
}
