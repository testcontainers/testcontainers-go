package clickhouse

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed mounts/zk_config.xml.tpl
var zookeeperConfigTpl string

const (
	defaultUser         = "default"
	defaultDatabaseName = "clickhouse"
)

const (
	// containerPorts {
	httpPort   = nat.Port("8123/tcp")
	nativePort = nat.Port("9000/tcp")
	// }
)

// ClickHouseContainer represents the ClickHouse container type used in the module
type ClickHouseContainer struct {
	testcontainers.Container
	DbName   string //nolint:staticcheck //FIXME
	User     string
	Password string
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

	connectionString := fmt.Sprintf("clickhouse://%s:%s@%s/%s%s", c.User, c.Password, host, c.DbName, extraArgs)
	return connectionString, nil
}

// ZookeeperOptions arguments for zookeeper in clickhouse
type ZookeeperOptions struct {
	Host, Port string
}

// renderZookeeperConfig generate default zookeeper configuration for clickhouse
func renderZookeeperConfig(settings ZookeeperOptions) ([]byte, error) {
	tpl, err := template.New("bootstrap.yaml").Parse(zookeeperConfigTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse zookeeper config file template: %w", err)
	}

	var bootstrapConfig bytes.Buffer
	if err := tpl.Execute(&bootstrapConfig, settings); err != nil {
		return nil, fmt.Errorf("failed to render zookeeper bootstrap config template: %w", err)
	}

	return bootstrapConfig.Bytes(), nil
}

// WithZookeeper pass a config to connect clickhouse with zookeeper and make clickhouse as cluster
func WithZookeeper(host, port string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		f, err := os.CreateTemp("", "clickhouse-tc-config-")
		if err != nil {
			return fmt.Errorf("temporary file: %w", err)
		}

		defer f.Close()

		// write data to the temporary file
		data, err := renderZookeeperConfig(ZookeeperOptions{Host: host, Port: port})
		if err != nil {
			return fmt.Errorf("zookeeper config: %w", err)
		}
		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("write zookeeper config: %w", err)
		}
		cf := testcontainers.ContainerFile{
			HostFilePath:      f.Name(),
			ContainerFilePath: "/etc/clickhouse-server/config.d/zookeeper_config.xml",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

// WithInitScripts sets the init scripts to be run when the container starts
func WithInitScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
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

		return nil
	}
}

// WithConfigFile sets the XML config file to be used for the clickhouse container
// It will also set the "configFile" parameter to the path of the config file
// as a command line argument to the container.
func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/clickhouse-server/config.d/config.xml",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

// WithConfigFile sets the YAML config file to be used for the clickhouse container
// It will also set the "configFile" parameter to the path of the config file
// as a command line argument to the container.
func WithYamlConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/clickhouse-server/config.d/config.yaml",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

// WithDatabase sets the initial database to be created when the container starts
// It can be used to define a different name for the default database that is created when the image is first started.
// If it is not specified, then the default value("clickhouse") will be used.
func WithDatabase(dbName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["CLICKHOUSE_DB"] = dbName

		return nil
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is required for you to use the ClickHouse image. It must not be empty or undefined.
// This environment variable sets the password for ClickHouse.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["CLICKHOUSE_PASSWORD"] = password

		return nil
	}
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a user and its password.
// It will create the specified user with superuser power.
// If it is not specified, then the default user of clickhouse will be used.
func WithUsername(user string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if user == "" {
			user = defaultUser
		}

		req.Env["CLICKHOUSE_USER"] = user

		return nil
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the ClickHouse container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ClickHouseContainer, error) {
	return Run(ctx, "clickhouse/clickhouse-server:23.3.8.21-alpine", opts...)
}

// Run creates an instance of the ClickHouse container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ClickHouseContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
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
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *ClickHouseContainer
	if container != nil {
		c = &ClickHouseContainer{Container: container}
	}
	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	c.User = req.Env["CLICKHOUSE_USER"]
	c.Password = req.Env["CLICKHOUSE_PASSWORD"]
	c.DbName = req.Env["CLICKHOUSE_DB"]

	return c, nil
}
