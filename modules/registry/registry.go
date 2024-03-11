package registry

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// RegistryContainer represents the Registry container type used in the module
type RegistryContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Registry container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RegistryContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "registry:2.8.3",
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

	return &RegistryContainer{Container: container}, nil
}
