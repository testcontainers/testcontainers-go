package timeplus

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// containerPorts {
	httpPort   = "8123/tcp"
	nativePort = "8463/tcp"
	// }
)

// Container represents the Timeplus container type used in the module.
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the Timeplus container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(httpPort, nativePort),
		testcontainers.WithWaitStrategyAndDeadline(
			300*time.Second,
			wait.ForHTTP("/ping").
				WithPort(httpPort).
				WithStatusCodeMatcher(func(status int) bool {
					return status == 200
				}).
				WithStartupTimeout(300*time.Second),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run timeplus: %w", err)
	}

	return c, nil
}

// HTTPEndpoint returns the HTTP endpoint of the Timeplus container for the
// ClickHouse-compatible HTTP API (port 8123).
// The returned URL has the form "http://host:port".
func (c *Container) HTTPEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, httpPort, "http")
	if err != nil {
		return "", fmt.Errorf("http endpoint: %w", err)
	}

	return endpoint, nil
}
