package clickhouse

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultUser         = "default"
	defaultDatabaseName = "clickhouse"
)

const defaultImage = "clickhouse/clickhouse-server:23.3.8.21-alpine"

const (
	// containerPorts {
	httpPort   = nat.Port("8123/tcp")
	nativePort = nat.Port("9000/tcp")
	// }
)

// ClickHouseContainer represents the ClickHouse container type used in the module
type ClickHouseContainer struct {
	testcontainers.Container
	dbName   string
	user     string
	password string
}

// ConnectionHost returns the host and port of the clickhouse container, using the default, native 9000 port, and
// obtaining the host and exposed port from the container
func (c *ClickHouseContainer) ConnectionHost(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, nativePort)
	if err != nil {
		return "", err
	}

	return host + ":" + port.Port(), nil
}

// ConnectionString returns the dsn string for the clickhouse container, using the default, native 9000 port, and
// obtaining the host and exposed port from the container. It also accepts a variadic list of extra arguments
// which will be appended to the dsn string. The format of the extra arguments is the same as the
// connection string format, e.g. "dial_timeout=300ms" or "skip_verify=false"
func (c *ClickHouseContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	host, err := c.ConnectionHost(ctx)
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

	connectionString := fmt.Sprintf("clickhouse://%s:%s@%s/%s%s", c.user, c.password, host, c.dbName, extraArgs)
	return connectionString, nil
}

// WithInitScripts sets the init scripts to be run when the container starts
func WithInitScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		initScripts := []testcontainers.ContainerFile{}
		for _, script := range scripts {
			cf := testcontainers.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
				FileMode:          0o755,
			}
			initScripts = append(initScripts, cf)
		}
		req.Files = append(req.Files, initScripts...)
	}
}

// WithConfigFile sets the XML config file to be used for the clickhouse container
// It will also set the "configFile" parameter to the path of the config file
// as a command line argument to the container.
func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/clickhouse-server/config.d/config.xml",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)
	}
}

// WithConfigFile sets the YAML config file to be used for the clickhouse container
// It will also set the "configFile" parameter to the path of the config file
// as a command line argument to the container.
func WithYamlConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/clickhouse-server/config.d/config.yaml",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)
	}
}

// WithDatabase sets the initial database to be created when the container starts
// It can be used to define a different name for the default database that is created when the image is first started.
// If it is not specified, then the default value("clickhouse") will be used.
func WithDatabase(dbName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["CLICKHOUSE_DB"] = dbName
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is required for you to use the ClickHouse image. It must not be empty or undefined.
// This environment variable sets the password for ClickHouse.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["CLICKHOUSE_PASSWORD"] = password
	}
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a user and its password.
// It will create the specified user with superuser power.
// If it is not specified, then the default user of clickhouse will be used.
func WithUsername(user string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		if user == "" {
			user = defaultUser
		}

		req.Env["CLICKHOUSE_USER"] = user
	}
}

// RunContainer creates an instance of the ClickHouse container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ClickHouseContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: defaultImage,
		Env: map[string]string{
			"CLICKHOUSE_USER":     defaultUser,
			"CLICKHOUSE_PASSWORD": defaultUser,
			"CLICKHOUSE_DB":       defaultDatabaseName,
		},
		ExposedPorts: []string{httpPort.Port(), nativePort.Port()},
		WaitingFor: wait.ForAll(
			wait.NewHTTPStrategy("/").WithPort(httpPort).WithStatusCodeMatcher(func(status int) bool {
				return status == 200
			}),
		),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	user := req.Env["CLICKHOUSE_USER"]
	password := req.Env["CLICKHOUSE_PASSWORD"]
	dbName := req.Env["CLICKHOUSE_DB"]

	return &ClickHouseContainer{Container: container, dbName: dbName, password: password, user: user}, nil
}
