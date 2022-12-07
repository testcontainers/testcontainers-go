package postgres

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// postgresContainer represents the postgres container type used in the module
type postgresContainer struct {
	testcontainers.Container
}

// setupPostgres creates an instance of the postgres container type
func setupPostgres(ctx context.Context) (*postgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "postgres:11-alpine",
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &postgresContainer{Container: container}, nil
}
