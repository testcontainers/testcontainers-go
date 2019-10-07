package canned

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	user       = "user"
	password   = "password"
	database   = "database"
	image      = "postgres"
	defaultTag = "11.5"
	port       = "5432/tcp"
)

type PostgreSQLContainerRequest struct {
	testcontainers.GenericContainerRequest
	User     string
	Password string
	Database string
}

type postgresqlContainer struct {
	Container testcontainers.Container
	db        *sql.DB
	req       PostgreSQLContainerRequest
}

// GetDriver returns a sql.DB connecting to the previously started Postgres DB.
// All the parameters are taken from the previous PostgreSQLContainerRequest
func (c *postgresqlContainer) GetDriver(ctx context.Context) (*sql.DB, error) {

	host, err := c.Container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := c.Container.MappedPort(ctx, port)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host,
		mappedPort.Int(),
		c.req.User,
		c.req.Password,
		c.req.Database,
	))
	if err != nil {
		return nil, err
	}

	return db, nil
}

// PostgreSQLContainer creates and (optionally) starts a Postgres database.
// If autostarted, the function returns only after a successful execution of a query
// (confirming that the database is ready)
func PostgreSQLContainer(ctx context.Context, req PostgreSQLContainerRequest) (*postgresqlContainer, error) {

	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}

	// With the current logic it's not really possible to allow other ports...
	req.ExposedPorts = []string{port}

	if req.Env == nil {
		req.Env = map[string]string{}
	}

	// Set the default values if none were provided in the request
	if req.Image == "" {
		req.Image = fmt.Sprintf("%s:%s", image, defaultTag)
	}

	if req.User == "" {
		req.User = user
	}

	if req.Password == "" {
		req.Password = password
	}

	if req.Database == "" {
		req.Database = database
	}

	req.Env["POSTGRES_USER"] = req.User
	req.Env["POSTGRES_PASSWORD"] = req.Password
	req.Env["POSTGRES_DB"] = req.Database

	connectorVars := map[string]interface{}{
		"port":     port,
		"user":     req.User,
		"password": req.Password,
		"database": req.Database,
	}

	req.WaitingFor = wait.ForSQL(postgresConnectorFromTarget, connectorVars)

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	postgresC := &postgresqlContainer{
		Container: c,
		req:       req,
	}

	if req.Started {
		if err := c.Start(ctx); err != nil {
			return postgresC, errors.Wrap(err, "failed to start container")
		}
	}

	return postgresC, nil
}

func postgresConnectorFromTarget(ctx context.Context, target wait.StrategyTarget, variables wait.SQLVariables) (driver.Connector, error) {

	host, err := target.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := target.MappedPort(ctx, nat.Port(variables["port"].(string)))
	if err != nil {
		return nil, err
	}

	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host,
		mappedPort.Int(),
		variables["user"],
		variables["password"],
		variables["database"],
	)

	return pq.NewConnector(connString)
}
