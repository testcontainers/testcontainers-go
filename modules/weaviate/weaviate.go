package weaviate

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	image    = "semitechnologies/weaviate:1.24.6"
	httpPort = "8080/tcp"
	grpcPort = "50051/tcp"
)

// WeaviateContainer represents the Weaviate container type used in the module
type WeaviateContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Weaviate container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*WeaviateContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        image,
		Cmd:          []string{"--host", "0.0.0.0", "--scheme", "http", "--port", "8080"},
		ExposedPorts: []string{httpPort, grpcPort},
		Env: map[string]string{
			"AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED": "true",
			"PERSISTENCE_DATA_PATH":                   "/var/lib/weaviate",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(httpPort).WithStartupTimeout(5*time.Second),
			wait.ForListeningPort(grpcPort).WithStartupTimeout(5*time.Second),
		),
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

	return &WeaviateContainer{Container: container}, nil
}

// HttpHostAddress returns the schema and host of the Weaviate container.
// At the moment, it only supports the http scheme.
func (c *WeaviateContainer) HttpHostAddress(ctx context.Context) (string, string, error) {
	port, err := c.MappedPort(ctx, httpPort)
	if err != nil {
		return "", "", fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get container host")
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
		return "", fmt.Errorf("failed to get container host")
	}

	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}
