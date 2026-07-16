package kurrentdb

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort  = "2113/tcp"
	defaultImage = "kurrentplatform/kurrentdb:latest"
)

// Container represents the KurrentDB container type used in the module.
type Container struct {
	testcontainers.Container
	insecure bool
}

// WithInsecure sets the container to run in insecure mode (no TLS).
// This is the default behaviour; the option exists for explicitness.
func WithInsecure() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["KURRENTDB__INSECURE"] = "true"
		return nil
	}
}

// Run creates an instance of the KurrentDB container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithEnv(map[string]string{
			"KURRENTDB__INSECURE": "true",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/health/live").
				WithPort(defaultPort).
				WithStatusCodeMatcher(func(status int) bool {
					return status == http.StatusNoContent
				}),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		// Inspect the environment to pick up the effective TLS setting after all options.
		insecure := true
		inspect, inspErr := ctr.Inspect(ctx)
		if inspErr == nil {
			for _, env := range inspect.Config.Env {
				if v, ok := strings.CutPrefix(env, "KURRENTDB__INSECURE="); ok {
					insecure = v == "true"
					break
				}
			}
		}
		c = &Container{
			Container: ctr,
			insecure:  insecure,
		}
	}

	if err != nil {
		return c, fmt.Errorf("run kurrentdb: %w", err)
	}

	return c, nil
}

// ConnectionString returns the connection string for the KurrentDB container.
// When the container is running in insecure mode (no TLS), the returned URL
// includes "?tls=false"; otherwise only the base URL is returned.
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("get host: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", fmt.Errorf("get mapped port: %w", err)
	}

	if c.insecure {
		return fmt.Sprintf("kurrentdb://%s:%s?tls=false", host, port.Port()), nil
	}

	return fmt.Sprintf("kurrentdb://%s:%s", host, port.Port()), nil
}
