package openfga

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Container represents the OpenFGA container type used in the module
type Container struct {
	*testcontainers.DockerContainer
}

// GrpcEndpoint returns the gRPC endpoint for the OpenFGA container,
// which uses the 8081/tcp port.
func (c *Container) GrpcEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "8081/tcp", "http")
}

// HttpEndpoint returns the HTTP endpoint for the OpenFGA container,
// which uses the 8080/tcp port.
func (c *Container) HttpEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "8080/tcp", "http")
}

// PlaygroundEndpoint returns the playground endpoint for the OpenFGA container,
// which is the HTTP endpoint with the path /playground in the port 3000/tcp.
func (c *Container) PlaygroundEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, "3000/tcp", "http")
	if err != nil {
		return "", fmt.Errorf("failed to get playground endpoint: %w", err)
	}

	return fmt.Sprintf("%s/playground", endpoint), nil
}

// Run creates an instance of the OpenFGA container type
func Run(ctx context.Context, img string, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	req := testcontainers.Request{
		Image:        img,
		Cmd:          []string{"run"},
		ExposedPorts: []string{"3000/tcp", "8080/tcp", "8081/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/healthz").WithPort("8080/tcp").WithResponseMatcher(func(r io.Reader) bool {
				bs, err := io.ReadAll(r)
				if err != nil {
					return false
				}

				return (strings.Contains(string(bs), "SERVING"))
			}),
			wait.ForHTTP("/playground").WithPort("3000/tcp").WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
		),
		Started: true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	ctr, err := testcontainers.Run(ctx, req)
	var c *Container
	if ctr != nil {
		c = &Container{DockerContainer: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
