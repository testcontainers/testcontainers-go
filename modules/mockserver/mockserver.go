package mockserver

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// MockServerContainer represents the MockServer container type used in the module
type MockServerContainer struct {
	testcontainers.Container
}

// Deprecated: use Run instead
// RunContainer creates an instance of the MockServer container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MockServerContainer, error) {
	return Run(ctx, "mockserver/mockserver:5.15.0", opts...)
}

// Run creates an instance of the MockServer container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MockServerContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("1080/tcp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("started on port: 1080"),
			wait.ForListeningPort("1080/tcp").SkipInternalCheck(),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *MockServerContainer
	if ctr != nil {
		c = &MockServerContainer{Container: ctr}
	}
	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}

// GetURL returns the URL of the MockServer container
func (c *MockServerContainer) URL(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := c.MappedPort(ctx, "1080/tcp")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s:%d", host, port.Int()), nil
}
