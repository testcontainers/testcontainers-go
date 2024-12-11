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
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"1080/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("started on port: 1080"),
			wait.ForListeningPort("1080/tcp").SkipInternalCheck(),
		),
		Env: map[string]string{},
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
	var c *MockServerContainer
	if container != nil {
		c = &MockServerContainer{Container: container}
	}
	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
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
