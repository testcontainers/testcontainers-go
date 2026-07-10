package cratedb

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// containerPorts {
	httpPort = "4200/tcp"
	pgPort   = "5432/tcp"
	// }

	defaultHeapSize = "512m"
)

// Container represents the CrateDB container type used in the module.
type Container struct {
	testcontainers.Container
}

// WithHeapSize sets the CRATE_HEAP_SIZE environment variable to configure the
// JVM heap size for the CrateDB container.
func WithHeapSize(size string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"CRATE_HEAP_SIZE": size,
	})
}

// Run creates an instance of the CrateDB container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(httpPort, pgPort),
		testcontainers.WithEnv(map[string]string{
			"CRATE_HEAP_SIZE": defaultHeapSize,
		}),
		testcontainers.WithWaitStrategy(
			wait.NewHTTPStrategy("/").WithPort(httpPort).WithStatusCodeMatcher(func(status int) bool {
				return status == 200
			}),
			wait.ForListeningPort(pgPort),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run cratedb: %w", err)
	}

	return c, nil
}

// HTTPEndpoint returns the HTTP endpoint of the CrateDB container, for the
// Admin UI and REST API on port 4200.
func (c *Container) HTTPEndpoint(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	port, err := c.MappedPort(ctx, httpPort)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

// PGConnectionString returns the PostgreSQL wire-protocol connection string for
// the CrateDB container. CrateDB's default user is "crate" with no password,
// and the default schema/database is "doc". It also accepts a variadic list of
// extra arguments which will be appended to the connection string as query
// parameters, e.g. "sslmode=disable" or "connect_timeout=10".
func (c *Container) PGConnectionString(ctx context.Context, args ...string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	port, err := c.MappedPort(ctx, pgPort)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	extraArgs := ""
	if len(args) > 0 {
		extraArgs = "?" + strings.Join(args, "&")
	}

	return fmt.Sprintf("postgres://crate@%s:%s/doc%s", host, port.Port(), extraArgs), nil
}
