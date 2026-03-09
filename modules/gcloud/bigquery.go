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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("9050/tcp", "9060/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/discovery/v1/apis/bigquery/v2/rest").WithPort("9050/tcp").WithStartupTimeout(time.Second * 5)),
	}

	moduleOpts = append(moduleOpts, opts...)

	settings, err := applyOptions(opts)
	if err != nil {
		return nil, err
	}

	moduleOpts = append(moduleOpts, testcontainers.WithCmdArgs("--project", settings.ProjectID))

	// Process data yaml file only for the BigQuery container.
	if settings.bigQueryDataYaml != nil {
		containerPath := "/testcontainers-data.yaml"

		moduleOpts = append(moduleOpts, testcontainers.WithCmdArgs("--data-from-yaml", containerPath))

		moduleOpts = append(moduleOpts, testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader:            settings.bigQueryDataYaml,
			ContainerFilePath: containerPath,
			FileMode:          0o644,
		}))
	}

	return newGCloudContainer(ctx, img, 9050, settings, "http", moduleOpts...)
}
