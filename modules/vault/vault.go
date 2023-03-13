package vault

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

// vaultContainer represents the vault container type used in the module
type vaultContainer struct {
	testcontainers.Container
}

// StartContainer creates an instance of the vault container type
func StartContainer(ctx context.Context, opts ...Option) (*vaultContainer, error) {
	config := &Config{
		imageName: "vault:latest",
		port:      8200,
	}

	for _, opt := range opts {
		opt(config)
	}

	req := testcontainers.ContainerRequest{
		Image:        config.imageName,
		ExposedPorts: []string{fmt.Sprintf("%d/tcp", config.port)},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &vaultContainer{Container: container}, nil
}
