package pinecone

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

// Container represents the Pinecone container type used in the module
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the Pinecone container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"5080/tcp"},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// HttpEndpoint returns the http endpoint for the pinecone container
//
//nolint:revive,staticcheck //FIXME
func (c *Container) HttpEndpoint() (string, error) {
	return c.PortEndpoint(context.Background(), "5080/tcp", "http")
}
