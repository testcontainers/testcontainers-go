package weaviate

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	httpPort = nat.Port("8080/tcp")
	grpcPort = nat.Port("50051/tcp")
)

// WeaviateContainer represents the Weaviate container type used in the module
type WeaviateContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Weaviate container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*WeaviateContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "semitechnologies/weaviate:1.24.5",
		Cmd:          []string{"--host", "0.0.0.0", "--scheme", "http", "--port", httpPort.Port()},
		ExposedPorts: []string{string(httpPort), string(grpcPort)},
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

func (c *WeaviateContainer) getHostAddress(ctx context.Context, port nat.Port) (string, error) {
	containerPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host")
	}

	return fmt.Sprintf("%s:%s", host, containerPort.Port()), nil
}

// HttpHostAddress returns the schema and host of the Weaviate container.
// At the moment, it only supports the http scheme.
func (c *WeaviateContainer) HttpHostAddress(ctx context.Context) (string, string, error) {
	httpHostAddress, err := c.getHostAddress(ctx, httpPort)
	if err != nil {
		return "", "", err
	}
	return "http", httpHostAddress, nil
}

// GrpcHostAddress returns the gRPC host of the Weaviate container.
// At the moment, it only supports unsecured gRPC connection.
func (c *WeaviateContainer) GrpcHostAddress(ctx context.Context) (string, error) {
	grpcHostAddress, err := c.getHostAddress(ctx, grpcPort)
	if err != nil {
		return "", err
	}
	return grpcHostAddress, nil
}
