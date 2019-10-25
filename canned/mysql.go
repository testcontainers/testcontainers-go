package canned

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

const (
	mysqlRootPassword = "root_password"
	mysqlUser         = "user"
	mysqlPassword     = "password"
	mysqlDatabase     = "database"
	mysqlImage        = "mysql"
	mysqlDefaultTag   = "8.0"
	mysqlPort         = "3306/tcp"
)

type MySQLContainerRequest struct {
	testcontainers.GenericContainerRequest
	RootPassword string
	Username     string
	Password     string
	Database     string
}

type MySQLContainer struct {
	Container testcontainers.Container
	db        *sql.DB
	req       MySQLContainerRequest
}

func (c *MySQLContainer) GetDriver(ctx context.Context) (*sql.DB, error) {

	host, err := c.Container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := c.Container.MappedPort(ctx, mysqlPort)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s",
		c.req.Username,
		c.req.Password,
		host,
		mappedPort.Int(),
		c.req.Database,
	))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewMySQLContainer(ctx context.Context, req MySQLContainerRequest) (*MySQLContainer, error) {

	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}

	// With the current logic it's not really possible to allow other ports...
	req.ExposedPorts = []string{mysqlPort}

	if req.Env == nil {
		req.Env = map[string]string{}
	}

	// Set the default values if none were provided in the request
	if req.Image == "" && req.FromDockerfile.Context == "" {
		req.Image = fmt.Sprintf("%s:%s", mysqlImage, mysqlDefaultTag)
	}

	if req.RootPassword == "" {
		req.RootPassword = mysqlRootPassword
	}

	if req.Username == "" {
		req.Username = mysqlUser
	}

	if req.Password == "" {
		req.Password = mysqlPassword
	}

	if req.Database == "" {
		req.Database = mysqlDatabase
	}

	req.Env["MYSQL_ROOT_PASSWORD"] = req.RootPassword
	req.Env["MYSQL_USERNAME"] = req.Username
	req.Env["MYSQL_PASSWORD"] = req.Password
	req.Env["MYSQL_DATABASE"] = req.Database

	req.WaitingFor = wait.ForLog("port: 3306  MySQL Community Server - GPL")

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	mysqlC := &MySQLContainer{
		Container: c,
		req:       req,
	}

	if req.Started {
		if err := c.Start(ctx); err != nil {
			return mysqlC, errors.Wrap(err, "failed to start container")
		}
	}

	return mysqlC, nil
}
