package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const defaultUser = "postgres"
const defaultPassword = "postgres"
const defaultPostgresImage = "docker.io/postgres:11-alpine"

// PostgresContainer represents the postgres container type used in the module
type PostgresContainer struct {
	testcontainers.Container
	dbName   string
	user     string
	password string
}

// ConnectionString returns the connection string for the postgres container, using the default 5432 port, and
// obtaining the host and exposed port from the container.
func (c *PostgresContainer) ConnectionString(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, containerPort.Port(), c.user, c.password, c.dbName)
	return connStr, nil
}

// PostgresContainerOption is a function that configures the postgres container, affecting the container request
type PostgresContainerOption func(req *testcontainers.ContainerRequest)

// WithWaitStrategy sets the wait strategy for the postgres container
func WithWaitStrategy(strategies ...wait.Strategy) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(1 * time.Minute)
	}
}

// WithImage sets the image to be used for the postgres container
func WithImage(image string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if image == "" {
			image = defaultPostgresImage
		}

		req.Image = image
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
		Image: defaultPostgresImage,
		Env: map[string]string{
			"POSTGRES_USER":     defaultUser,
			"POSTGRES_PASSWORD": defaultPassword,
			"POSTGRES_DB":       defaultUser, // defaults to the user name
		},
		ExposedPorts: []string{"5432/tcp"},
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

	user := req.Env["POSTGRES_USER"]
	password := req.Env["POSTGRES_PASSWORD"]
	dbName := req.Env["POSTGRES_DB"]

	return &PostgresContainer{Container: container, dbName: dbName, password: password, user: user}, nil
}
