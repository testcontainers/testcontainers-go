package weaviate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	httpPort = "8080/tcp"
	grpcPort = "50051/tcp"
)

// WeaviateContainer represents the Weaviate container type used in the module
type WeaviateContainer struct {
	testcontainers.Container
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Weaviate container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*WeaviateContainer, error) {
	return Run(ctx, "semitechnologies/weaviate:1.29.0", opts...)
}

// Run creates an instance of the Weaviate container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*WeaviateContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		Cmd:          []string{"--host", "0.0.0.0", "--scheme", "http", "--port", "8080"},
		ExposedPorts: []string{httpPort, grpcPort},
		Env: map[string]string{
			"AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED": "true",
			"PERSISTENCE_DATA_PATH":                   "/var/lib/weaviate",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(httpPort).WithStartupTimeout(5*time.Second),
			wait.ForListeningPort(grpcPort).WithStartupTimeout(5*time.Second),
			wait.ForHTTP("/v1/.well-known/ready").WithPort(httpPort),
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
	var c *WeaviateContainer
	if container != nil {
		c = &WeaviateContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// HttpHostAddress returns the schema and host of the Weaviate container.
// At the moment, it only supports the http scheme.
//
//nolint:revive,staticcheck //FIXME
func (c *WeaviateContainer) HttpHostAddress(ctx context.Context) (string, string, error) {
	port, err := c.MappedPort(ctx, httpPort)
	if err != nil {
		return "", "", fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", "", errors.New("failed to get container host")
	}

	return "http", fmt.Sprintf("%s:%s", host, port.Port()), nil
}

// GrpcHostAddress returns the gRPC host of the Weaviate container.
// At the moment, it only supports unsecured gRPC connection.
func (c *WeaviateContainer) GrpcHostAddress(ctx context.Context) (string, error) {
	port, err := c.MappedPort(ctx, grpcPort)
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", errors.New("failed to get container host")
	}

	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}
