package arangodb

import (
	"context"
	"fmt"
	"net/http"

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
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{defaultPort},
		Env: map[string]string{
			"ARANGO_ROOT_PASSWORD": defaultPassword,
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	// Wait for the container to be ready once we know the credentials
	genericContainerReq.WaitingFor = wait.ForAll(
		wait.ForListeningPort(defaultPort),
		wait.ForHTTP("/_admin/status").
			WithPort(defaultPort).
			WithBasicAuth(DefaultUser, req.Env["ARANGO_ROOT_PASSWORD"]).
			WithHeaders(map[string]string{
				"Accept": "application/json",
			}).
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
	)

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{Container: container, password: req.Env["ARANGO_ROOT_PASSWORD"]}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
