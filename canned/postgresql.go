package canned

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

const (
	user       = "user"
	password   = "password"
	database   = "database"
	image      = "postgres"
	defaultTag = "11.5"
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
	ctx       context.Context
}

func (c *postgresqlContainer) GetDriver() (*sql.DB, error) {

	host, err := c.Container.Host(c.ctx)
	if err != nil {
		return nil, err
	}

	port, err := c.Container.MappedPort(c.ctx, "5432/tcp")
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host,
		port.Int(),
		c.req.User,
		c.req.Password,
		c.req.Database,
	))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func PostgreSQLContainer(ctx context.Context, req PostgreSQLContainerRequest) (*postgresqlContainer, error) {

	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}

	req.ExposedPorts = []string{"3306/tcp"}
	req.Env = map[string]string{}
	req.Started = true

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

	req.WaitingFor = wait.ForLog("database system is ready to accept connections")

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	postgresC := &postgresqlContainer{
		Container: c,
		req:       req,
		ctx:       ctx,
	}

	if req.Started {
		if err := c.Start(ctx); err != nil {
			return postgresC, errors.Wrap(err, "failed to start container")
		}
	}

	return postgresC, nil
}
