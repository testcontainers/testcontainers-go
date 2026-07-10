package orientdb

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// defaultBinaryPort is the OrientDB binary/remote protocol port (used by Java and JDBC clients).
	defaultBinaryPort = "2424/tcp"
	// defaultHTTPPort is the OrientDB HTTP / Studio UI port.
	defaultHTTPPort = "2480/tcp"

	defaultRootPassword = "rootpwd"
)

// Container represents the OrientDB container type used in the module.
type Container struct {
	testcontainers.Container
	rootPassword string
}

// Run creates an instance of the OrientDB container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithEnv(map[string]string{
			"ORIENTDB_ROOT_PASSWORD": defaultRootPassword,
		}),
		testcontainers.WithExposedPorts(defaultBinaryPort, defaultHTTPPort),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/").
				WithPort(defaultHTTPPort).
				WithStatusCodeMatcher(func(status int) bool {
					return status == http.StatusOK
				}),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, rootPassword: defaultRootPassword}
	}

	if err != nil {
		return c, fmt.Errorf("run orientdb: %w", err)
	}

	// Extract the actual root password from the container environment so that
	// it reflects any password set via WithRootPassword.
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect orientdb: %w", err)
	}

	for _, env := range inspect.Config.Env {
		if value, ok := strings.CutPrefix(env, "ORIENTDB_ROOT_PASSWORD="); ok {
			c.rootPassword = value
			break
		}
	}

	return c, nil
}

// ServerURL returns the OrientDB binary remote protocol URL for Java/JDBC clients,
// in the format "remote:<host>:<port>".
func (c *Container) ServerURL(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("orientdb server url: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultBinaryPort)
	if err != nil {
		return "", fmt.Errorf("orientdb server url: %w", err)
	}

	return fmt.Sprintf("remote:%s:%s", host, port.Port()), nil
}

// StudioURL returns the OrientDB Studio web UI URL, in the format "http://<host>:<port>".
func (c *Container) StudioURL(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, defaultHTTPPort, "http")
	if err != nil {
		return "", fmt.Errorf("orientdb studio url: %w", err)
	}

	return endpoint, nil
}

// WithRootPassword sets the root password for the OrientDB container via the
// ORIENTDB_ROOT_PASSWORD environment variable.
func WithRootPassword(password string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"ORIENTDB_ROOT_PASSWORD": password,
	})
}
