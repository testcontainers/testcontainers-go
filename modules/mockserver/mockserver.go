package mockserver

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// defaultImage is the default MockServer container image
const defaultImage = "mockserver/mockserver"

// MockServerContainer represents the MockServer container type used in the module
type MockServerContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the MockServer container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MockServerContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultImage,
		ExposedPorts: []string{"1080/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("started on port: 1080"),
			wait.ForListeningPort("1080/tcp"),
		),
		Env: map[string]string{},
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

	return &MockServerContainer{Container: container}, nil
}

// GetHost returns the host on which the MockServer container is listening
func (c *MockServerContainer) GetHost(ctx context.Context) (string, error) {
	return c.Host(ctx)
}

// GetPort returns the port on which the MockServer container is listening
func (c *MockServerContainer) GetPort(ctx context.Context) (int, error) {
	port, err := c.MappedPort(ctx, "1080/tcp")
	if err != nil {
		return 0, err
	}
	return port.Int(), nil
}
