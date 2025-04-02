package arangodb

import (
	"context"
	"fmt"
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort     = "8529/tcp"
	defaultUser     = "root"
	defaultPassword = "root"
)

// Container represents the ArangoDB container type used in the module
type Container struct {
	testcontainers.Container
}

// TransportAddress returns the transport address of the ArangoDB container
func (c *Container) TransportAddress(ctx context.Context) (string, error) {
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
			"ARANGO_ROOT_USERNAME": defaultUser,
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
			WithBasicAuth(defaultUser, req.Env["ARANGO_ROOT_PASSWORD"]).
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
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
