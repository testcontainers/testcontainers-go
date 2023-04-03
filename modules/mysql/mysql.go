package mysql

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

// RunContainer creates an instance of the MySQL container type
func RunContainer(ctx context.Context, opts ...testcontainers.CustomizeRequestOption) (*MySQLContainer, error) {
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

	opts = append(opts, func(req *testcontainers.GenericContainerRequest) {
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

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}
	for _, opt := range opts {
		opt(&genericContainerReq)
	}

	username, ok := req.Env["MYSQL_USER"]
	if !ok {
		username = rootUser
	}
	password := req.Env["MYSQL_PASSWORD"]

	if len(password) == 0 && password == "" && !strings.EqualFold(rootUser, username) {
		return nil, fmt.Errorf("empty password can be used only with the root user")
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
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

func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MYSQL_USER"] = username
	}
}

func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MYSQL_PASSWORD"] = password
	}
}

func WithDatabase(database string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MYSQL_DATABASE"] = database
	}
}

func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/mysql/conf.d/my.cnf",
			FileMode:          0755,
		}
		req.Files = append(req.Files, cf)
	}
}

func WithScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
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
