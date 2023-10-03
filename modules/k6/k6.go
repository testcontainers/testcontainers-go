package k6

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// K6Container represents the K6 container type used in the module
type K6Container struct {
	testcontainers.Container
}

// RunContainer creates an instance of the K6 container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*K6Container, error) {
	req := testcontainers.ContainerRequest{
		Image: "szkiba/k6x",
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

	return &K6Container{Container: container}, nil
}
