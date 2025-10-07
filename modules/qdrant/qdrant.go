package qdrant

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// QdrantContainer represents the Qdrant container type used in the module
type QdrantContainer struct {
	testcontainers.Container
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Qdrant container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*QdrantContainer, error) {
	return Run(ctx, "qdrant/qdrant:v1.7.4", opts...)
}

// Run creates an instance of the Qdrant container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*QdrantContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("6333/tcp", "6334/tcp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort("6333/tcp").WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("6334/tcp").WithStartupTimeout(5*time.Second),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *QdrantContainer
	if ctr != nil {
		c = &QdrantContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run qdrant: %w", err)
	}

	return c, nil
}

// RESTEndpoint returns the REST endpoint of the Qdrant container
func (c *QdrantContainer) RESTEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "6333/tcp", "http")
}

// GRPCEndpoint returns the gRPC endpoint of the Qdrant container
func (c *QdrantContainer) GRPCEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "6334/tcp", "")
}

// WebUI returns the web UI endpoint of the Qdrant container
func (c *QdrantContainer) WebUI(ctx context.Context) (string, error) {
	s, err := c.RESTEndpoint(ctx)
	if err != nil {
		return "", err
	}

	return s + "/dashboard", nil
}
