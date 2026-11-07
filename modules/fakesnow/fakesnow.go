// Package fakesnow provides a Testcontainers module for fakesnow
// (https://github.com/tekumara/fakesnow), a Snowflake fake backed by DuckDB.
// It runs as an HTTP server speaking the Snowflake REST API, so the official
// gosnowflake driver connects to it unmodified.
package fakesnow

import (
	"context"
	"fmt"

	sf "github.com/snowflakedb/gosnowflake"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// defaultPort is the port fakesnow serves the Snowflake REST API on.
const defaultPort = "64616/tcp"

// Container represents the fakesnow container type used in the module.
type Container struct {
	testcontainers.Container
	opts options
}

// ConnectionString returns a gosnowflake DSN for the container, for use with
// sql.Open("snowflake", dsn).
//
// The account is always part of the DSN. gosnowflake otherwise derives it from
// the host, which has no account part when the host is a container address.
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("get host: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", fmt.Errorf("get mapped port: %w", err)
	}

	dsn, err := sf.DSN(&sf.Config{
		Account:  c.opts.account,
		User:     c.opts.user,
		Password: c.opts.password,
		Host:     host,
		Port:     int(port.Num()),
		Protocol: "http",
		Database: c.opts.database,
		Schema:   c.opts.schema,
	})
	if err != nil {
		return "", fmt.Errorf("build dsn: %w", err)
	}

	return dsn, nil
}

// Run creates an instance of the fakesnow container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	settings := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&settings); err != nil {
				return nil, fmt.Errorf("fakesnow option: %w", err)
			}
		}
	}

	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultPort),
		// The log line is what fakesnow prints once it is serving. The port on
		// its own is not enough: docker publishes it before the application
		// binds, so a connection can still be refused after it looks ready.
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("Application startup complete"),
			wait.ForListeningPort(defaultPort),
		)),
	)
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, opts: settings}
	}

	if err != nil {
		return c, fmt.Errorf("run fakesnow: %w", err)
	}

	return c, nil
}
