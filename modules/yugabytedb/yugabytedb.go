package yugabytedb

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yugabyte/gocql"
)

const (
	mastDashboardPort    = "7000"
	tServerDashboardPort = "9000"
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

// YugabyteDBContainer represents the yugabyteDB container type used in the module
type YugabyteDBContainer struct {
	testcontainers.Container

	ysqlDatabaseName     string
	ysqlDatabaseUser     string
	ysqlDatabasePassword string
}

// Run creates an instance of the yugabyteDB container type and automatically starts it.
// A default configuration is used for the container, but it can be customized using the
// provided options.
// When using default configuration values it is recommended to use the provided
// [*YugabyteDBContainer.YSQLConnectionString] and [*YugabyteDBContainer.YCQLConfigureClusterConfig]
// methods to use the container in their respective clients.
//
// # Example using YSQL dialect:
//
//	ctx := context.Background()
//	y, err := yugabytedb.Run(ctx, "yugabytedb/yugabyte")
//	if err != nil {
//		return err
//	}
//	defer func() {
//		if err := testcontainers.TerminateContainer(y); err != nil {
//			return err
//		}
//	}()
//
//	connStr, err := y.YSQLConnectionString(ctx)
//	if err != nil {
//		return err
//	}
//
//	db, err := sql.Open("postgres", connStr)
//	if err != nil {
//		return err
//	}
//	defer db.Close()
//
// # Example using YCQL dialect:
//
//	ctx := context.Background()
//	y, err := yugabytedb.Run(ctx, "yugabytedb/yugabyte")
//	if err != nil {
//		return err
//	}
//	defer func() {
//		if err := testcontainers.TerminateContainer(y); err != nil {
//			return err
//		}
//	}()
//
//	cluster := gocql.NewCluster()
//	y.YCQLConfigureClusterConfig(ctx, cluster)
//
//	session, err := cluster.CreateSession()
//	if err != nil {
//		return err
//	}
//	defer session.Close()
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*YugabyteDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
		Cmd:   []string{"bin/yugabyted", "start", "--background=false"},
		WaitingFor: wait.ForAll(
			wait.ForLog("YugabyteDB Started").WithOccurrence(1),
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
	var c *YugabyteDBContainer
	if container != nil {
		c = &YugabyteDBContainer{
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
//
// # Example:
//
//	connStr, err := y.YSQLConnectionString(ctx, "sslmode=disable")
//	if err != nil {
//		return err
//	}
//
// The above code will return a connection string with the sslmode set to disable.
func (y *YugabyteDBContainer) YSQLConnectionString(ctx context.Context, args ...string) (string, error) {
	host, err := y.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := y.MappedPort(ctx, ysqlPort)
	if err != nil {
		return "", err
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

// YCQLConfigureClusterConfig merges the provider [gocql.ClusterConfig] with the
// yugabyteDB container configuration.
// The method sets the hosts, keyspace and authenticator fields of the cluster
// configuration to the values of the yugabyteDB container.
//
// # Example:
//
//	cluster := gocql.NewCluster()
//	ctr.YCQLConfigureClusterConfig(ctx, cluster)
//	session, err := cluster.CreateSession()
//	if err != nil {
//		return err
//	}
//	defer session.Close()
//
// [gocql.ClusterConfig]: https://pkg.go.dev/github.com/gocql/gocql?tab=doc#ClusterConfig
func (y *YugabyteDBContainer) YCQLConfigureClusterConfig(ctx context.Context, cfg *gocql.ClusterConfig) error {
	host, err := y.Host(ctx)
	if err != nil {
		return err
	}

	mappedPort, err := y.MappedPort(ctx, ycqlPort)
	if err != nil {
		return err
	}

	inspect, err := y.Container.Inspect(ctx)
	if err != nil {
		return err
	}

	var (
		keyspace string
		user     string
		password string
	)

	for _, env := range inspect.Config.Env {
		parts := strings.SplitN(env, "=", 2)
		switch parts[0] {
		case ycqlKeyspaceEnv:
			keyspace = parts[1]
		case ycqlUserNameEnv:
			user = parts[1]
		case ycqlPasswordEnv:
			password = parts[1]
		}
	}

	cfg.Hosts = append(cfg.Hosts, net.JoinHostPort(host, mappedPort.Port()))
	cfg.Keyspace = keyspace
	cfg.Authenticator = gocql.PasswordAuthenticator{
		Username: user,
		Password: password,
	}

	return nil
}
