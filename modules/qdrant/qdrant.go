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
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"6333/tcp", "6334/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("6333/tcp").WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("6334/tcp").WithStartupTimeout(5*time.Second),
		),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *QdrantContainer
	if container != nil {
		c = &QdrantContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// RESTEndpoint returns the REST endpoint of the Qdrant container
func (c *QdrantContainer) RESTEndpoint(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "6333/tcp")
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host")
	}

	return fmt.Sprintf("http://%s:%s", host, containerPort.Port()), nil
}

// GRPCEndpoint returns the gRPC endpoint of the Qdrant container
func (c *QdrantContainer) GRPCEndpoint(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "6334/tcp")
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host")
	}

	return fmt.Sprintf("%s:%s", host, containerPort.Port()), nil
}

// WebUI returns the web UI endpoint of the Qdrant container
func (c *QdrantContainer) WebUI(ctx context.Context) (string, error) {
	s, err := c.RESTEndpoint(ctx)
	if err != nil {
		return "", err
	}

	return s + "/dashboard", nil
}
