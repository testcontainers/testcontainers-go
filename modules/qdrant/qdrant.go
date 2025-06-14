package qdrant

import (
	"context"
	"errors"
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
	modulesOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("6333/tcp", "6334/tcp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort("6333/tcp").WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("6334/tcp").WithStartupTimeout(5*time.Second),
		)),
	}

	modulesOpts = append(modulesOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, modulesOpts...)
	var c *QdrantContainer
	if ctr != nil {
		c = &QdrantContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
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
		return "", errors.New("failed to get container host")
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
		return "", errors.New("failed to get container host")
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
