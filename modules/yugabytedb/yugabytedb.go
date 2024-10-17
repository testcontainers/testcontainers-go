package yugabytedb

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	ycqlPort = "9042/tcp"

	ycqlKeyspaceEnv = "YCQL_KEYSPACE"
	ycqlUserNameEnv = "YCQL_USER"
	ycqlPasswordEnv = "YCQL_PASSWORD"

	ycqlKeyspace = "yugabyte"
	ycqlUserName = "yugabyte"
	ycqlPassword = "yugabyte"
)

const (
	ysqlPort = "5433/tcp"

	ysqlDatabaseNameEnv     = "YSQL_DB"
	ysqlDatabaseUserEnv     = "YSQL_USER"
	ysqlDatabasePasswordEnv = "YSQL_PASSWORD"

	ysqlDatabaseName     = "yugabyte"
	ysqlDatabaseUser     = "yugabyte"
	ysqlDatabasePassword = "yugabyte"
)

// Container represents the yugabyteDB container type used in the module
type Container struct {
	testcontainers.Container

	ysqlDatabaseName     string
	ysqlDatabaseUser     string
	ysqlDatabasePassword string
}

// Run creates an instance of the yugabyteDB container type and automatically starts it.
// A default configuration is used for the container, but it can be customized using the
// provided options.
// When using default configuration values it is recommended to use the provided
// [*Container.YSQLConnectionString] and [*Container.YCQLConfigureClusterConfig]
// methods to use the container in their respective clients.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
		Cmd:   []string{"bin/yugabyted", "start", "--background=false"},
		WaitingFor: wait.ForAll(
			wait.ForLog("YugabyteDB Started").WithOccurrence(1),
			wait.ForLog("Data placement constraint successfully verified").WithOccurrence(1),
			wait.ForListeningPort(ysqlPort),
			wait.ForListeningPort(ycqlPort),
		),
		ExposedPorts: []string{ycqlPort, ysqlPort},
		Env: map[string]string{
			ycqlKeyspaceEnv:         ycqlKeyspace,
			ycqlUserNameEnv:         ycqlUserName,
			ycqlPasswordEnv:         ycqlPassword,
			ysqlDatabaseNameEnv:     ysqlDatabaseName,
			ysqlDatabaseUserEnv:     ysqlDatabaseUser,
			ysqlDatabasePasswordEnv: ysqlDatabasePassword,
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{
			Container:            container,
			ysqlDatabaseName:     req.Env[ysqlDatabaseNameEnv],
			ysqlDatabaseUser:     req.Env[ysqlDatabaseUserEnv],
			ysqlDatabasePassword: req.Env[ysqlDatabasePasswordEnv],
		}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// YSQLConnectionString returns a connection string for the yugabyteDB container
// using the configured database name, user, password, port, host and additional
// arguments.
// Additional arguments are appended to the connection string as query parameters
// in the form of key=value pairs separated by "&".
func (y *Container) YSQLConnectionString(ctx context.Context, args ...string) (string, error) {
	host, err := y.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	mappedPort, err := y.MappedPort(ctx, ysqlPort)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?%s",
		y.ysqlDatabaseUser,
		y.ysqlDatabasePassword,
		net.JoinHostPort(host, mappedPort.Port()),
		y.ysqlDatabaseName,
		strings.Join(args, "&"),
	), nil
}
