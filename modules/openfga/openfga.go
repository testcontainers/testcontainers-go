package openfga

import (
	"context"
	"io"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// OpenFGAContainer represents the OpenFGA container type used in the module
type OpenFGAContainer struct {
	testcontainers.Container
}

func (c *OpenFGAContainer) GrpcEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "8081/tcp", "http")
}

func (c *OpenFGAContainer) HttpEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "8080/tcp", "http")
}

// RunContainer creates an instance of the OpenFGA container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenFGAContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "openfga/openfga:v1.5.0",
		Cmd:          []string{"run"},
		ExposedPorts: []string{"3000/tcp", "8080/tcp", "8081/tcp"},
		WaitingFor: wait.ForHTTP("/healthz").WithPort("8080/tcp").WithResponseMatcher(func(r io.Reader) bool {
			bs, err := io.ReadAll(r)
			if err != nil {
				return false
			}

			return (strings.Contains(string(bs), "SERVING"))
		}),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &OpenFGAContainer{Container: container}, nil
}
