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
	config *Config
}

// StartContainer creates an instance of the MySQL container type
func StartContainer(ctx context.Context, image string, opts ...Option) (*MySQLContainer, error) {
	config := &Config{
		username: defaultUser,
		password: defaultPassword,
		database: defaultDatabaseName,
	}

	for _, opt := range opts {
		opt(config)
	}

	mysqlEnv := map[string]string{}
	mysqlEnv["MYSQL_DATABASE"] = config.database
	if !strings.EqualFold(rootUser, config.username) {
		mysqlEnv["MYSQL_USER"] = config.username
	}
	if len(config.password) != 0 && config.password != "" {
		mysqlEnv["MYSQL_PASSWORD"] = config.password
		mysqlEnv["MYSQL_ROOT_PASSWORD"] = config.password
	} else if strings.EqualFold(rootUser, config.username) {
		mysqlEnv["MYSQL_ALLOW_EMPTY_PASSWORD"] = "yes"
	} else {
		return nil, fmt.Errorf("empty password can be used only with the root user")
	}

	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env:          mysqlEnv,
		WaitingFor:   wait.ForLog("port: 3306  MySQL Community Server - GPL"),
	}

	//if config.configFile != "" {
	//	req.Files = []testcontainers.ContainerFile{
	//		{HostFilePath: config.configFile, ContainerFilePath: "/etc/mysql/conf.d/my.cnf", FileMode: 700},
	//	}
	//}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &MySQLContainer{container, config}, nil
}

func (c *MySQLContainer) Username() string {
	return c.config.username
}

func (c *MySQLContainer) Password() string {
	return c.config.password
}

func (c *MySQLContainer) Database() string {
	return c.config.database
}
