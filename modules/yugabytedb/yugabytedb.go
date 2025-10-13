package yugabytedb

import (
	"context"
	"fmt"
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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithCmd("bin/yugabyted", "start", "--background=false"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("YugabyteDB Started").WithOccurrence(1),
			wait.ForLog("Data placement constraint successfully verified").WithOccurrence(1),
			wait.ForListeningPort(ysqlPort),
			wait.ForListeningPort(ycqlPort),
		),
		testcontainers.WithExposedPorts(ycqlPort, ysqlPort),
		testcontainers.WithEnv(map[string]string{
			ycqlKeyspaceEnv:         ycqlKeyspace,
			ycqlUserNameEnv:         ycqlUserName,
			ycqlPasswordEnv:         ycqlPassword,
			ysqlDatabaseNameEnv:     ysqlDatabaseName,
			ysqlDatabaseUserEnv:     ysqlDatabaseUser,
			ysqlDatabasePasswordEnv: ysqlDatabasePassword,
		}),
	}

	ctr, err := testcontainers.Run(ctx, img, append(moduleOpts, opts...)...)
	var c *Container
	if ctr != nil {
		c = &Container{
			Container:            ctr,
			ysqlDatabaseName:     ysqlDatabaseName,
			ysqlDatabaseUser:     ysqlDatabaseUser,
			ysqlDatabasePassword: ysqlDatabasePassword,
		}
	}

	if err != nil {
		return c, fmt.Errorf("run yugabytedb: %w", err)
	}

	// Inspect the container to get the actual env var values after user customizations
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect yugabytedb: %w", err)
	}

	var foundName, foundUser, foundPassword bool
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, ysqlDatabaseNameEnv+"="); ok {
			c.ysqlDatabaseName, foundName = v, true
		}
		if v, ok := strings.CutPrefix(env, ysqlDatabaseUserEnv+"="); ok {
			c.ysqlDatabaseUser, foundUser = v, true
		}
		if v, ok := strings.CutPrefix(env, ysqlDatabasePasswordEnv+"="); ok {
			c.ysqlDatabasePassword, foundPassword = v, true
		}

		if foundName && foundUser && foundPassword {
			break
		}
	}

	return c, nil
}

// YSQLConnectionString returns a connection string for the yugabyteDB container
// using the configured database name, user, password, port, host and additional
// arguments.
// Additional arguments are appended to the connection string as query parameters
// in the form of key=value pairs separated by "&".
func (y *Container) YSQLConnectionString(ctx context.Context, args ...string) (string, error) {
	endpoint, err := y.PortEndpoint(ctx, ysqlPort, "")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?%s",
		y.ysqlDatabaseUser,
		y.ysqlDatabasePassword,
		endpoint,
		y.ysqlDatabaseName,
		strings.Join(args, "&"),
	), nil
}
