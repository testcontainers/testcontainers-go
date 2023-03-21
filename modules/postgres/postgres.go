package postgres

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer represents the postgres container type used in the module
type PostgresContainer struct {
	testcontainers.Container
}

// PostgresContainerOption is a function that configures the postgres container, affecting the container request
type PostgresContainerOption func(req *testcontainers.ContainerRequest)

// WithWaitStrategy sets the wait strategy for the postgres container
func WithWaitStrategy(strategies ...wait.Strategy) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(1 * time.Minute)
	}
}

// WithInitialDatabase sets the initial database to be created when the container starts, including the user and password
func WithInitialDatabase(user string, password string, dbName string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["POSTGRES_USER"] = user
		req.Env["POSTGRES_PASSWORD"] = password
		req.Env["POSTGRES_DB"] = dbName
	}
}

// WithInitDBArgs sets the initdb arguments for the postgres container.
// The value is a space separated string of arguments as postgres initdb would expect them
func WithInitDBArgs(args string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["POSTGRES_INITDB_ARGS"] = args
	}
}

// StartContainer creates an instance of the postgres container type
func StartContainer(ctx context.Context, opts ...PostgresContainerOption) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:11-alpine",
		Env:          map[string]string{},
		ExposedPorts: []string{},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
	}

	for _, opt := range opts {
		opt(&req)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{Container: container}, nil
}
