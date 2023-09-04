package elasticsearch

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

const (
	minimalImageVersion = "8.0.0"
)

// ElasticsearchContainer represents the Elasticsearch container type used in the module
type ElasticsearchContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Elasticsearch container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ElasticsearchContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: minimalImageVersion,
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

	return &ElasticsearchContainer{Container: container}, nil
}
