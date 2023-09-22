package gcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// BigQueryContainer represents the GCloud container type used in the module for BigQuery
type BigQueryContainer struct {
	testcontainers.Container
	Settings options
	URI      string
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
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "ghcr.io/goccy/bigquery-emulator:0.4.3",
			ExposedPorts: []string{"9050/tcp", "9060/tcp"},
			WaitingFor:   wait.ForHTTP("/discovery/v1/apis/bigquery/v2/rest").WithPort("9050/tcp").WithStartupTimeout(time.Second * 5),
		},
		Started: true,
	}

	settings := applyOptions(req, opts)

	req.Cmd = []string{"--project", settings.ProjectID}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	bigQueryContainer := BigQueryContainer{
		Container: container,
		Settings:  settings,
	}

	uri, err := containerURI(ctx, &bigQueryContainer)
	if err != nil {
		return nil, err
	}

	bigQueryContainer.URI = uri

	return &bigQueryContainer, nil
}
