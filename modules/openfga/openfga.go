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
//
//nolint:revive,staticcheck //FIXME
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

	return endpoint + "/playground", nil
}

// Deprecated: use Run instead
// RunContainer creates an instance of the OpenFGA container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenFGAContainer, error) {
	return Run(ctx, "openfga/openfga:v1.5.0", opts...)
}

// Run creates an instance of the OpenFGA container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OpenFGAContainer, error) {
	modulesOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithCmd("run"),
		testcontainers.WithExposedPorts("3000/tcp", "8080/tcp", "8081/tcp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
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
		)),
	}

	modulesOpts = append(modulesOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, modulesOpts...)
	var c *OpenFGAContainer
	if ctr != nil {
		c = &OpenFGAContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}
