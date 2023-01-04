package mongodb

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// mongodbContainer represents the mongodb container type used in the module
type mongodbContainer struct {
	testcontainers.Container
}

// setupMongodb creates an instance of the mongodb container type
func setupMongodb(ctx context.Context) (*mongodbContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "mongo:6",
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &mongodbContainer{Container: container}, nil
}
