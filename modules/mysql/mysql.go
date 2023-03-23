package mysql

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"path/filepath"
	"strings"
)

const rootUser = "root"
const defaultUser = "test"
const defaultPassword = "test"
const defaultDatabaseName = "test"
const defaultImage = "mysql:8"

// MySQLContainer represents the MySQL container type used in the module
type MySQLContainer struct {
	testcontainers.Container
	username string
	password string
	database string
}

type MySQLContainerOption func(req *testcontainers.ContainerRequest)

// StartContainer creates an instance of the MySQL container type
func StartContainer(ctx context.Context, opts ...MySQLContainerOption) (*MySQLContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultImage,
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

func (c *MySQLContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	containerPort, err := c.MappedPort(ctx, "3306/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	extraArgs := ""
	if len(args) > 0 {
		extraArgs = strings.Join(args, "&")
	}
	if extraArgs != "" {
		extraArgs = "?" + extraArgs
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", c.username, c.password, host, containerPort.Port(), c.database, extraArgs)
	return connectionString, nil
}

// WithImage sets the image to be used for the mysql container
func WithImage(image string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if image == "" {
			image = "mysql:8"
		}

		req.Image = image
	}
}

func WithUsername(username string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["MYSQL_USER"] = username
	}
}

func WithPassword(password string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["MYSQL_PASSWORD"] = password
	}
}

func WithDatabase(database string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["MYSQL_DATABASE"] = database
	}
}

func WithConfigFile(configFile string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/mysql/conf.d/my.cnf",
			FileMode:          0755,
		}
		req.Files = append(req.Files, cf)
	}
}

func WithScripts(scripts ...string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		var initScripts []testcontainers.ContainerFile
		for _, script := range scripts {
			cf := testcontainers.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
				FileMode:          0755,
			}
			initScripts = append(initScripts, cf)
		}
		req.Files = append(req.Files, initScripts...)
	}
}
