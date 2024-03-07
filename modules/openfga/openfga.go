package openfga

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// OpenFGAContainer represents the OpenFGA container type used in the module
type OpenFGAContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the OpenFGA container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenFGAContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "openfga/openfga:v1.5.0",
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

	return &OpenFGAContainer{Container: container}, nil
}
