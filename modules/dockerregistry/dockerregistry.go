package dockerregistry

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// DockerRegistryContainer represents the DockerRegistry container type used in the module
type DockerRegistryContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the DockerRegistry container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DockerRegistryContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "docker.io/registry:latest",
		ExposedPorts: []string{"5000/tcp"},
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

	return &DockerRegistryContainer{Container: container}, nil
}
