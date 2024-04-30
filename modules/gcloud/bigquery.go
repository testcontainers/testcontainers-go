package gcloud

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RunBigQueryContainer creates an instance of the GCloud container type for BigQuery.
// The URI will always use http:// as the protocol.
func RunBigQueryContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "ghcr.io/goccy/bigquery-emulator:0.6.1",
			ExposedPorts: []string{"9050/tcp", "9060/tcp"},
			WaitingFor:   wait.ForHTTP("/discovery/v1/apis/bigquery/v2/rest").WithPort("9050/tcp").WithStartupTimeout(time.Second * 5),
		},
		Started: true,
	}

	settings, err := applyOptions(&req, opts)
	if err != nil {
		return nil, err
	}

	req.Cmd = []string{"--project", settings.ProjectID}

	for _, opt := range opts {
		opt.Customize(&req)
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	bigqueryContainer, err := newGCloudContainer(ctx, 9050, container, settings)
	if err != nil {
		return nil, err
	}

	// always prepend http:// to the URI
	bigqueryContainer.URI = "http://" + bigqueryContainer.URI

	return bigqueryContainer, nil
}
