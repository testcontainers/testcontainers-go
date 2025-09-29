package arangodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort = "8529/tcp"

	// DefaultUser is the default username for the ArangoDB container.
	// This is the username to be used when connecting to the ArangoDB instance.
	DefaultUser = "root"

	defaultPassword = "root"
)

// Container represents the ArangoDB container type used in the module
type Container struct {
	testcontainers.Container
	password string
}

// Credentials returns the credentials for the ArangoDB container:
// first return value is the username, second is the password.
func (c *Container) Credentials() (string, string) {
	return DefaultUser, c.password
}

// HTTPEndpoint returns the HTTP endpoint of the ArangoDB container, using the following format: `http://$host:$port`.
func (c *Container) HTTPEndpoint(ctx context.Context) (string, error) {
	hostPort, err := c.PortEndpoint(ctx, defaultPort, "http")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return hostPort, nil
}

// Run creates an instance of the ArangoDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithEnv(map[string]string{
			"ARANGO_ROOT_PASSWORD": defaultPassword,
		}),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(defaultPort)),
	}

	moduleOpts = append(moduleOpts, opts...)

	// configure the wait strategy after all the options have been applied
	moduleOpts = append(moduleOpts, withWaitStrategy())

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if container != nil {
		c = &Container{Container: container, password: defaultPassword}
	}

	if err != nil {
		return c, fmt.Errorf("run arangodb: %w", err)
	}

	inspect, err := container.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect arangodb: %w", err)
	}

	for _, env := range inspect.Config.Env {
		if strings.HasPrefix(env, "ARANGO_ROOT_PASSWORD=") {
			c.password = strings.TrimPrefix(env, "ARANGO_ROOT_PASSWORD=")
			break
		}
	}

	return c, nil
}
