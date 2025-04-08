package toxiproxy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	ControlPort      = "8474/tcp"
	firstProxiedPort = 8666
	defaultPortRange = 31
)

// Container represents the Toxiproxy container type used in the module
type Container struct {
	testcontainers.Container
}

// URI returns the URI of the Toxiproxy container
func (c *Container) URI(ctx context.Context) (string, error) {
	portEndpoint, err := c.PortEndpoint(ctx, ControlPort, "http")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return portEndpoint, nil
}

// Run creates an instance of the Toxiproxy container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{ControlPort},
		WaitingFor: wait.ForHTTP("/version").WithPort(ControlPort).WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusOK
		}),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, fmt.Errorf("apply: %w", err)
			}
		}
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	// Expose the ports in the range, starting from the first proxied port
	portsInRange := make([]string, 0, settings.portRange)
	for i := range settings.portRange {
		portsInRange = append(portsInRange, fmt.Sprintf("%d/tcp", firstProxiedPort+i))
	}
	genericContainerReq.ExposedPorts = append(genericContainerReq.ExposedPorts, portsInRange...)

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
