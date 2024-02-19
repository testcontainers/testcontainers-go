package weaviate

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// WeaviateContainer represents the Weaviate container type used in the module
type WeaviateContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Weaviate container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*WeaviateContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "semitechnologies/weaviate:1.23.9",
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

	return &WeaviateContainer{Container: container}, nil
}
