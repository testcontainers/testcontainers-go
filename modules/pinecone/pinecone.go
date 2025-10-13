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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("5080/tcp"),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run pinecone: %w", err)
	}

	return c, nil
}

// HttpEndpoint returns the http endpoint for the pinecone container
//
//nolint:revive,staticcheck //FIXME
func (c *Container) HttpEndpoint() (string, error) {
	return c.PortEndpoint(context.Background(), "5080/tcp", "http")
}
