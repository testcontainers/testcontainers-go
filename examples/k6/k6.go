package k6

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// k6Container represents the k6 container type used in the module
type k6Container struct {
	testcontainers.Container
}

// runContainer creates an instance of the k6 container type
func runContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*k6Container, error) {
	req := testcontainers.ContainerRequest{
		Image: "grafana/k6",
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

	return &k6Container{Container: container}, nil
}
