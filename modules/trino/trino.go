package trino

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
	// httpPort is the default port exposed by the Trino coordinator.
	httpPort = "8080/tcp"
)

// Container represents the Trino container type used in the module.
type Container struct {
	testcontainers.Container
}

// ConnectionString returns the HTTP connection string for the Trino coordinator,
// e.g. "http://localhost:8080".
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, httpPort, "http")
}

// Run creates an instance of the Trino container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(httpPort),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/v1/info").
				WithPort(httpPort).
				WithResponseMatcher(isTrinoReady).
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
		return c, fmt.Errorf("run trino: %w", err)
	}

	return c, nil
}

// isTrinoReady checks the /v1/info response body and returns true once the
// coordinator reports starting=false.
func isTrinoReady(body io.Reader) bool {
	var info map[string]any
	if err := json.NewDecoder(body).Decode(&info); err != nil {
		return false
	}
	starting, _ := info["starting"].(bool)
	return !starting
}
