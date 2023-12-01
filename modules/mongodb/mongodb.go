package mongodb

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// defaultImage is the default MongoDB container image
const defaultImage = "mongo:6"

// MongoDBContainer represents the MongoDB container type used in the module
type MongoDBContainer struct {
	testcontainers.Container
	username string
	password string
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
		Env: map[string]string{},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}
	username := req.Env["MONGO_INITDB_ROOT_USERNAME"]
	password := req.Env["MONGO_INITDB_ROOT_PASSWORD"]
	if username != "" && password == "" || username == "" && password != "" {
		return nil, fmt.Errorf("if you specify username or password, you must provide both of them")
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	if username != "" && password != "" {
		return &MongoDBContainer{Container: container, username: username, password: password}, nil
	}
	return &MongoDBContainer{Container: container}, nil
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a username and its password.
// It will create the specified user with superuser power.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MONGO_INITDB_ROOT_USERNAME"] = username
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is used in conjunction with WithUsername to set a username and its password.
// It will set the superuser password for MongoDB.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MONGO_INITDB_ROOT_PASSWORD"] = password
	}
}

// ConnectionString returns the connection string for the MongoDB container.
// If you provide a username and a password, the connection string will also include them.
func (c *MongoDBContainer) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := c.MappedPort(ctx, "27017/tcp")
	if err != nil {
		return "", err
	}
	if c.username != "" && c.password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s", c.username, c.password, host, port.Port()), nil
	}
	return c.Endpoint(ctx, "mongodb")
}
