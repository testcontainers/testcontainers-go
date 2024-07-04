package gcloud

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RunBigQuery creates an instance of the GCloud container type for BigQuery.
// The URI will always use http:// as the protocol.
func RunBigQuery(ctx context.Context, img string, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	req := testcontainers.Request{
		Image:        img,
		ExposedPorts: []string{"9050/tcp", "9060/tcp"},
		WaitingFor:   wait.ForHTTP("/discovery/v1/apis/bigquery/v2/rest").WithPort("9050/tcp").WithStartupTimeout(time.Second * 5),
		Started:      true,
	}

	settings, err := applyOptions(&req, opts)
	if err != nil {
		return nil, err
	}

	req.Cmd = []string{"--project", settings.ProjectID}

	ctr, err := testcontainers.Run(ctx, req)
	if err != nil {
		return nil, err
	}

	bigQueryContainer, err := newGCloudContainer(ctx, 9050, ctr, settings)
	if err != nil {
		return nil, err
	}

	// always prepend http:// to the URI
	bigQueryContainer.URI = "http://" + bigQueryContainer.URI

	return bigQueryContainer, nil
}
