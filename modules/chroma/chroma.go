package chroma

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// ChromaContainer represents the Chroma container type used in the module
type ChromaContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Chroma container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ChromaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "chromadb/chroma:0.4.22.dev44",
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

	return &ChromaContainer{Container: container}, nil
}
