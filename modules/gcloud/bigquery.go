package gcloud

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Deprecated: use [bigquery.Run] instead.
// RunBigQueryContainer creates an instance of the GCloud container type for BigQuery.
func RunBigQueryContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	return RunBigQuery(ctx, "ghcr.io/goccy/bigquery-emulator:0.6.1", opts...)
}

// Deprecated: use [bigquery.Run] instead.
// RunBigQuery creates an instance of the GCloud container type for BigQuery.
// The URI uses http:// as the protocol.
func RunBigQuery(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GCloudContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        img,
			ExposedPorts: []string{"9050/tcp", "9060/tcp"},
			WaitingFor:   wait.ForHTTP("/discovery/v1/apis/bigquery/v2/rest").WithPort("9050/tcp").WithStartupTimeout(time.Second * 5),
		},
		Started: true,
	}

	settings, err := applyOptions(&req, opts)
	if err != nil {
		return nil, err
	}

	req.Cmd = append(req.Cmd, "--project", settings.ProjectID)

	// Process data yaml file only for the BigQuery container.
	if settings.bigQueryDataYaml != nil {
		containerPath := "/testcontainers-data.yaml"

		req.Cmd = append(req.Cmd, "--data-from-yaml", containerPath)

		req.Files = append(req.Files, testcontainers.ContainerFile{
			Reader:            settings.bigQueryDataYaml,
			ContainerFilePath: containerPath,
			FileMode:          0o644,
		})
	}

	return newGCloudContainer(ctx, req, 9050, settings, "http://")
}
