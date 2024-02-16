package qdrant

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// QdrantContainer represents the Qdrant container type used in the module
type QdrantContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Qdrant container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*QdrantContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "qdrant/qdrant:v1.7.4",
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

	return &QdrantContainer{Container: container}, nil
}
