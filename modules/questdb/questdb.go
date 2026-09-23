package questdb

import (
	"context"
	"fmt"
	"net"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// defaultHTTPPort is the port for the HTTP/Web Console and REST API.
	defaultHTTPPort = "9000/tcp"
	// defaultILPPort is the port for the InfluxDB line protocol ingestion.
	defaultILPPort = "9009/tcp"
	// defaultPGPort is the port for the PostgreSQL wire protocol.
	defaultPGPort = "8812/tcp"

	// defaultAdminUser is the built-in QuestDB admin username.
	defaultAdminUser = "admin"
	// defaultAdminPassword is the built-in QuestDB admin password.
	defaultAdminPassword = "quest"
	// defaultDatabase is the built-in QuestDB database name.
	defaultDatabase = "qdb"
)

// Container represents the QuestDB container type used in the module.
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the QuestDB container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultHTTPPort, defaultILPPort, defaultPGPort),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/").WithPort(defaultHTTPPort).WithStatusCodeMatcher(func(status int) bool {
				return status == 200
			}),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run questdb: %w", err)
	}

	return c, nil
}

// HTTPEndpoint returns the HTTP endpoint (Web Console and REST API) of the QuestDB container.
// The returned URL has the format "http://host:port".
func (c *Container) HTTPEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, defaultHTTPPort, "http")
}

// PGEndpoint returns the PostgreSQL wire protocol connection string for the QuestDB container.
// The returned URL has the format "postgres://admin:[REDACTED]@host:port/qdb".
func (c *Container) PGEndpoint(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("pg endpoint host: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultPGPort)
	if err != nil {
		return "", fmt.Errorf("pg endpoint port: %w", err)
	}

	return fmt.Sprintf("postgres://%s:%s@%s/%s",
		defaultAdminUser, defaultAdminPassword,
		net.JoinHostPort(host, port.Port()),
		defaultDatabase,
	), nil
}

// InfluxDBEndpoint returns the InfluxDB line protocol endpoint of the QuestDB container.
// The returned address has the format "host:port".
func (c *Container) InfluxDBEndpoint(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("influxdb endpoint host: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultILPPort)
	if err != nil {
		return "", fmt.Errorf("influxdb endpoint port: %w", err)
	}

	return net.JoinHostPort(host, port.Port()), nil
}
