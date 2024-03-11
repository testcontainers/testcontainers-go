package registry

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RegistryContainer represents the Registry container type used in the module
type RegistryContainer struct {
	testcontainers.Container
}

// Address returns the address of the Registry container, using the HTTP protocol
func (c *RegistryContainer) Address(ctx context.Context) (string, error) {
	port, err := c.MappedPort(ctx, "5000")
	if err != nil {
		return "", err
	}

	ipAddress, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", ipAddress, port.Port()), nil
}

// RunContainer creates an instance of the Registry container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RegistryContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "registry:2.8.3",
		ExposedPorts: []string{"5000/tcp"},
		Env:          map[string]string{},
		WaitingFor:   wait.ForExposedPort(),
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

	return &RegistryContainer{Container: container}, nil
}
