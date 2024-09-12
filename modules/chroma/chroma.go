package chroma

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ChromaContainer represents the Chroma container type used in the module
type ChromaContainer struct {
	testcontainers.Container
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Chroma container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ChromaContainer, error) {
	return Run(ctx, "chromadb/chroma:0.4.24", opts...)
}

// Run creates an instance of the Chroma container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ChromaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"8000/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("8000/tcp"),
			wait.ForLog("Application startup complete"),
			wait.ForHTTP("/api/v1/heartbeat").WithStatusCodeMatcher(func(status int) bool {
				return status == 200
			}),
		), // 5 seconds it's not enough for the container to start
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
	var c *ChromaContainer
	if container != nil {
		c = &ChromaContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// RESTEndpoint returns the REST endpoint of the Chroma container
func (c *ChromaContainer) RESTEndpoint(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "8000/tcp")
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host")
	}

	return fmt.Sprintf("http://%s:%s", host, containerPort.Port()), nil
}
