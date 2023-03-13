package vault

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// vaultContainer represents the vault container type used in the module
type vaultContainer struct {
	testcontainers.Container
}

// StartContainer creates an instance of the vault container type
func StartContainer(ctx context.Context) (*vaultContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "vault:latest",
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
