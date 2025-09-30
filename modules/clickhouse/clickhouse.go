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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(httpPort.Port(), nativePort.Port()),
		testcontainers.WithEnv(map[string]string{
			"CLICKHOUSE_USER":     defaultUser,
			"CLICKHOUSE_PASSWORD": defaultUser,
			"CLICKHOUSE_DB":       defaultDatabaseName,
		}),
		testcontainers.WithWaitStrategy(wait.NewHTTPStrategy("/").WithPort(httpPort).WithStatusCodeMatcher(func(status int) bool {
			return status == 200
		},
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *ClickHouseContainer
	if ctr != nil {
		c = &ClickHouseContainer{Container: ctr}
	}
	if err != nil {
		return c, fmt.Errorf("run clickhouse: %w", err)
	}

	// initialize the credentials
	c.User = defaultUser
	c.Password = defaultUser
	c.DbName = defaultDatabaseName

	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect clickhouse: %w", err)
	}

	// refresh the credentials from the environment variables
	foundUser, foundPass, foundDB := false, false, false
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "CLICKHOUSE_USER="); ok {
			c.User, foundUser = v, true
		}
		if v, ok := strings.CutPrefix(env, "CLICKHOUSE_PASSWORD="); ok {
			c.Password, foundPass = v, true
		}
		if v, ok := strings.CutPrefix(env, "CLICKHOUSE_DB="); ok {
			c.DbName, foundDB = v, true
		}

		if foundUser && foundPass && foundDB {
			break
		}
	}

	return c, nil
}
