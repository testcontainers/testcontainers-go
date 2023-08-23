package mongodb

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// defaultImage is the default MongoDB container image
const defaultImage = "mongo:6"

// MongoDBContainer represents the MongoDB container type used in the module
type MongoDBContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the MongoDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MongoDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultImage,
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Waiting for connections"),
			wait.ForListeningPort("27017/tcp"),
		),
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

	return &MongoDBContainer{Container: container}, nil
}

// ConnectionString returns the connection string for the MongoDB container
func (c *MongoDBContainer) ConnectionString(ctx context.Context) (string, error) {
	return c.Endpoint(ctx, "mongodb")
}
