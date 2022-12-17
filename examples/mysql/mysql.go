package mysql

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// mysqlContainer represents the mysql container type used in the module
type mysqlContainer struct {
	testcontainers.Container
}

// setupMysql creates an instance of the mysql container type
func setupMysql(ctx context.Context) (*mysqlContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "mysql:8",
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &mysqlContainer{Container: container}, nil
}
