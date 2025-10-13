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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("8000/tcp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort("8000/tcp"),
			wait.ForLog("Application startup complete"),
			wait.ForHTTP("/api/v1/heartbeat").WithStatusCodeMatcher(func(status int) bool {
				return status == 200
			}),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *ChromaContainer
	if ctr != nil {
		c = &ChromaContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run chroma: %w", err)
	}

	return c, nil
}

// RESTEndpoint returns the REST endpoint of the Chroma container
func (c *ChromaContainer) RESTEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "8000/tcp", "http")
}
