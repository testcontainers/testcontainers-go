package presto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// containerPorts {
	httpPort = "8080/tcp"
	// }
)

// Container represents the Presto container type used in the module
type Container struct {
	testcontainers.Container
}

// ConnectionString returns the HTTP connection string for the Presto container,
// e.g. "http://localhost:8080".
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("get host: %w", err)
	}

	port, err := c.MappedPort(ctx, httpPort)
	if err != nil {
		return "", fmt.Errorf("get mapped port: %w", err)
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

// Run creates an instance of the Presto container type.
// It waits for the coordinator to be fully started by polling /v1/info until
// the "starting" field is false.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(httpPort),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/v1/info").
				WithPort(httpPort).
				WithResponseMatcher(func(body io.Reader) bool {
					var info map[string]any
					if err := json.NewDecoder(body).Decode(&info); err != nil {
						return false
					}
					starting, _ := info["starting"].(bool)
					return !starting
				}).
				WithStartupTimeout(2*time.Minute),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run presto: %w", err)
	}

	return c, nil
}
