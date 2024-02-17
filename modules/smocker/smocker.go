package smocker

import (
	"context"
	"fmt"
	"net"

	"github.com/testcontainers/testcontainers-go"
)

// SmockerContainer represents the Smocker container type used in the module
type SmockerContainer struct {
	testcontainers.Container
}

// ApiURL returns the URL of the Smocker API
func (c *SmockerContainer) ApiURL(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "8081/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	connStr := fmt.Sprintf("http://%s", net.JoinHostPort(host, containerPort.Port()))
	return connStr, nil
}

// MockURL returns the URL of the Smocker mock server
func (c *SmockerContainer) MockURL(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "8080/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	connStr := fmt.Sprintf("http://%s", net.JoinHostPort(host, containerPort.Port()))
	return connStr, nil
}

// RunContainer creates an instance of the Smocker container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*SmockerContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "thiht/smocker:0.18.5",
		ExposedPorts: []string{"8080/tcp", "8081/tcp"},
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

	return &SmockerContainer{Container: container}, nil
}
