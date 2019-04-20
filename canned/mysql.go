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

type MySQLContainerRequest struct {
	testcontainers.GenericContainerRequest
	RootPassword string
	Username     string
	Password     string
	Database     string
}

type mysqlContainer struct {
	Container testcontainers.Container
	db        *sql.DB
	req       MySQLContainerRequest
	ctx       context.Context
}

func (c *mysqlContainer) GetDriver(username, password, database string) (*sql.DB, error) {
	ip, err := c.Container.Host(c.ctx)
	if err != nil {
		return nil, err
	}
	port, err := c.Container.MappedPort(c.ctx, "3306/tcp")
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		username,
		password,
		ip,
		port.Port(),
		database,
	))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func MySQLContainer(ctx context.Context, req MySQLContainerRequest) (*mysqlContainer, error) {
	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}

	req.Image = "mysql:8.0"
	req.ExposedPorts = []string{"3306/tcp"}
	req.Env = map[string]string{}
	req.Started = true
	if req.RootPassword != "" {
		req.Env["MYSQL_ROOT_PASSWORD"] = req.RootPassword
	}
	if req.Username != "" {
		req.Env["MYSQL_USERNAME"] = req.Username
	}
	if req.Password != "" {
		req.Env["MYSQL_PASSWORD"] = req.Password
	}
	if req.Database != "" {
		req.Env["MYSQL_DATABASE"] = req.Database
	}

	req.WaitingFor = wait.ForLog("port: 3306  MySQL Community Server - GPL")

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}
	mysqlC := &mysqlContainer{
		Container: c,
	}
	mysqlC.req = req
	mysqlC.ctx = ctx

	if req.Started {
		if err := c.Start(ctx); err != nil {
			return mysqlC, errors.Wrap(err, "failed to start container")
		}
	}

	return mysqlC, nil
}
