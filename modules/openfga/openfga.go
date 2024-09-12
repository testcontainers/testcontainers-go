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

// OpenFGAContainer represents the OpenFGA container type used in the module
type OpenFGAContainer struct {
	testcontainers.Container
}

// GrpcEndpoint returns the gRPC endpoint for the OpenFGA container,
// which uses the 8081/tcp port.
func (c *OpenFGAContainer) GrpcEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "8081/tcp", "http")
}

// HttpEndpoint returns the HTTP endpoint for the OpenFGA container,
// which uses the 8080/tcp port.
func (c *OpenFGAContainer) HttpEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "8080/tcp", "http")
}

// PlaygroundEndpoint returns the playground endpoint for the OpenFGA container,
// which is the HTTP endpoint with the path /playground in the port 3000/tcp.
func (c *OpenFGAContainer) PlaygroundEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, "3000/tcp", "http")
	if err != nil {
		return "", fmt.Errorf("failed to get playground endpoint: %w", err)
	}

	return fmt.Sprintf("%s/playground", endpoint), nil
}

// Deprecated: use Run instead
// RunContainer creates an instance of the OpenFGA container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenFGAContainer, error) {
	return Run(ctx, "openfga/openfga:v1.5.0", opts...)
}

// Run creates an instance of the OpenFGA container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OpenFGAContainer, error) {
	req := testcontainers.ContainerRequest{
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

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *OpenFGAContainer
	if container != nil {
		c = &OpenFGAContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
