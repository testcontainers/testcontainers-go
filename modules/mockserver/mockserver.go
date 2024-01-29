package mockserver

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// defaultImage is the default MockServer container image
const defaultImage = "mockserver/mockserver:5.15.0"

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
