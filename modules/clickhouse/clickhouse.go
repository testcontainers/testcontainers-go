package clickhouse

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
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

// Deprecated: use Run instead
// RunContainer creates an instance of the ClickHouse container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ClickHouseContainer, error) {
	return Run(ctx, "clickhouse/clickhouse-server:23.3.8.21-alpine", opts...)
}

// Run creates an instance of the ClickHouse container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ClickHouseContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(httpPort.Port(), nativePort.Port()),
		testcontainers.WithWaitStrategy(wait.NewHTTPStrategy("/").WithPort(httpPort).WithStatusCodeMatcher(func(status int) bool {
			return status == 200
		})),
	}

	moduleOpts = append(moduleOpts, opts...)

	defaultSettings := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&defaultSettings); err != nil {
				return nil, fmt.Errorf("clickhouse option: %w", err)
			}
		}
	}

	// module options take precedence over default options
	moduleOpts = append(moduleOpts, testcontainers.WithEnv(defaultSettings.env), testcontainers.WithFiles(defaultSettings.files...))

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *ClickHouseContainer
	if container != nil {
		c = &ClickHouseContainer{
			Container: container,
			User:      defaultSettings.env["CLICKHOUSE_USER"],
			Password:  defaultSettings.env["CLICKHOUSE_PASSWORD"],
			DbName:    defaultSettings.env["CLICKHOUSE_DB"],
		}
	}
	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
