package milvus

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// MilvusContainer represents the Milvus container type used in the module
type MilvusContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Milvus container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MilvusContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "milvusdb/milvus:v2.3.9",
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

	return &MilvusContainer{Container: container}, nil
}
