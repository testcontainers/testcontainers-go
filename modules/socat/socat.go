package socat

import (
	"context"
	"fmt"
	"net/url"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
)

// Container represents the Socat container type used in the module.
// A socat container is used as a TCP proxy, enabling any TCP port
// of another container to be exposed publicly, even if that container
// does not make the port public itself.
type Container struct {
	testcontainers.Container
	targetURLs map[int]*url.URL
}

// Run creates an instance of the Socat container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      img,
			Entrypoint: []string{"/bin/sh"},
		},
		Started: true,
	}

	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, err
			}
		}
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	for k := range settings.targets {
		req.ExposedPorts = append(req.ExposedPorts, fmt.Sprintf("%d/tcp", k))
	}

	if settings.targetsCmd != "" {
		req.Cmd = append(req.Cmd, "-c", settings.targetsCmd)
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	targetURLs := map[int]*url.URL{}
	for k := range settings.targets {
		hostPort, err := c.PortEndpoint(ctx, nat.Port(fmt.Sprintf("%d/tcp", k)), "http")
		if err != nil {
			return c, fmt.Errorf("mapped port: %w", err)
		}

		targetURL, err := url.Parse(hostPort)
		if err != nil {
			return c, fmt.Errorf("url parse: %w", err)
		}
		targetURLs[k] = targetURL
	}

	c.targetURLs = targetURLs

	return c, nil
}

// TargetURL returns the URL for the exposed port of a target, nil if the port is not mapped
func (c *Container) TargetURL(exposedPort int) *url.URL {
	return c.targetURLs[exposedPort]
}
