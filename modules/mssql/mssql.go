package mssql

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// MSSQLServerContainer represents the MSSQLServer container type used in the module
type MSSQLServerContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the MSSQLServer container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MSSQLServerContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "mcr.microsoft.com/mssql/server:2022-latest",
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

	return &MSSQLServerContainer{Container: container}, nil
}
