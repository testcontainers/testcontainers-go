package tidb

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultUser     = "root"
	defaultPassword = ""
	defaultDatabase = "test"
	defaultPort     = "4000/tcp"
	restAPIPort     = "10080/tcp"
)

// Container represents the TiDB container type used in the module
type Container struct {
	testcontainers.Container
	username string
	password string
	database string
}

// ConnectionString returns a DSN connection string for the TiDB container,
// using the MySQL driver format. It is possible to pass extra parameters
// to the connection string, e.g. "tls=skip-verify".
func (c *Container) ConnectionString(ctx context.Context, args ...string) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, defaultPort, "")
	if err != nil {
		return "", err
	}

	extraArgs := ""
	if len(args) > 0 {
		extraArgs = "?" + strings.Join(args, "&")
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s%s", c.username, c.password, endpoint, c.database, extraArgs)
	return connectionString, nil
}

// MustConnectionString panics if the connection string cannot be determined.
func (c *Container) MustConnectionString(ctx context.Context, args ...string) string {
	addr, err := c.ConnectionString(ctx, args...)
	if err != nil {
		panic(err)
	}
	return addr
}

// Run creates an instance of the TiDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultPort, restAPIPort),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForListeningPort(defaultPort),
				wait.ForHTTP("/status").WithPort(restAPIPort),
			),
		),
	)
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{
			Container: ctr,
			username:  defaultUser,
			password:  defaultPassword,
			database:  defaultDatabase,
		}
	}

	if err != nil {
		return c, fmt.Errorf("run tidb: %w", err)
	}

	return c, nil
}
