package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

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
	return c.PortEndpoint(ctx, nativePort, "")
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
