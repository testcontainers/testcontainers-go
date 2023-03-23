package mysql

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"strings"
)

const rootUser = "root"
const defaultUser = "test"
const defaultPassword = "test"
const defaultDatabaseName = "test"

// MySQLContainer represents the MySQL container type used in the module
type MySQLContainer struct {
	testcontainers.Container
	username string
	password string
	database string
}

type MySQLContainerOption func(req *testcontainers.ContainerRequest)

// StartContainer creates an instance of the MySQL container type
func StartContainer(ctx context.Context, image string, opts ...PostgresContainerOption) (*MySQLContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_USER":     defaultUser,
			"MYSQL_PASSWORD": defaultPassword,
			"MYSQL_DATABASE": defaultDatabaseName,
		},
		WaitingFor: wait.ForLog("port: 3306  MySQL Community Server"),
	}

	opts = append(opts, func(req *testcontainers.ContainerRequest) {
		username := req.Env["MYSQL_USER"]
		password := req.Env["MYSQL_PASSWORD"]
		if strings.EqualFold(rootUser, username) {
			delete(req.Env, "MYSQL_USER")
		}
		if len(password) != 0 && password != "" {
			req.Env["MYSQL_ROOT_PASSWORD"] = password
		} else if strings.EqualFold(rootUser, username) {
			req.Env["MYSQL_ALLOW_EMPTY_PASSWORD"] = "yes"
			delete(req.Env, "MYSQL_PASSWORD")
		}
	})

	for _, opt := range opts {
		opt(&req)
	}

	username, ok := req.Env["MYSQL_USER"]
	if !ok {
		username = rootUser
	}
	password := req.Env["MYSQL_PASSWORD"]

	if len(password) == 0 && password == "" && !strings.EqualFold(rootUser, username) {
		return nil, fmt.Errorf("empty password can be used only with the root user")
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	database := req.Env["MYSQL_DATABASE"]

	return &MySQLContainer{container, username, password, database}, nil
}

func (c *MySQLContainer) Username() string {
	return c.username
}

func (c *MySQLContainer) Password() string {
	return c.password
}

func (c *MySQLContainer) Database() string {
	return c.database
}
