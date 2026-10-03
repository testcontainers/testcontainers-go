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
// e.g. "http://localhost:8080" (IPv6-safe).
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, httpPort, "http")
	if err != nil {
		return "", fmt.Errorf("get endpoint: %w", err)
	}

	return endpoint, nil
}

// Run creates an instance of the Presto container type.
// It waits for the coordinator to be fully started by:
//  1. Polling /v1/info until starting=false (catalogs loaded).
//  2. Polling /v1/cluster until at least one worker is active and can accept queries.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(httpPort),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				// Stage 1: coordinator has finished loading catalogs.
				wait.ForHTTP("/v1/info").
					WithPort(httpPort).
					WithResponseMatcher(func(body io.Reader) bool {
						var info map[string]any
						if err := json.NewDecoder(body).Decode(&info); err != nil {
							return false
						}
						starting, _ := info["starting"].(bool)
						return !starting
					}),
				// Stage 2: at least one worker has registered and can execute queries.
				wait.ForHTTP("/v1/cluster").
					WithPort(httpPort).
					WithResponseMatcher(func(body io.Reader) bool {
						var cluster struct {
							ActiveWorkers int `json:"activeWorkers"`
						}
						if err := json.NewDecoder(body).Decode(&cluster); err != nil {
							return false
						}
						return cluster.ActiveWorkers > 0
					}),
			).WithDeadline(2*time.Minute),
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
