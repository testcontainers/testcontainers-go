package toxiproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// ControlPort is the port of the Toxiproxy control API
	ControlPort = "8474/tcp"

	// firstProxiedPort is the first port of the range of ports that will be proxied
	firstProxiedPort = 8666
)

// Container represents the Toxiproxy container type used in the module
type Container struct {
	testcontainers.Container

	// proxiedEndpoints is a map of the proxied endpoints of the Toxiproxy container
	proxiedEndpoints map[int]string
}

// ProxiedEndpoint returns the endpoint for the proxied port in the Toxiproxy container,
// an error in case the port has no proxied endpoint.
func (c *Container) ProxiedEndpoint(p int) (string, string, error) {
	endpoint, ok := c.proxiedEndpoints[p]
	if !ok {
		return "", "", errors.New("port not found")
	}

	return net.SplitHostPort(endpoint)
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

	// Expose the ports for the proxies, starting from the first proxied port
	portsInRange := make([]string, 0, len(settings.proxies))
	for i, proxy := range settings.proxies {
		proxiedPort := firstProxiedPort + i
		// Update the listen port of the proxy
		proxy.Listen = fmt.Sprintf("0.0.0.0:%d", proxiedPort)
		portsInRange = append(portsInRange, fmt.Sprintf("%d/tcp", proxiedPort))
	}
	genericContainerReq.ExposedPorts = append(genericContainerReq.ExposedPorts, portsInRange...)

	// Render the config file
	jsonData, err := json.MarshalIndent(settings.proxies, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	// Apply the config file to the container with the proxies.
	if len(settings.proxies) > 0 {
		genericContainerReq.Files = append(genericContainerReq.Files, testcontainers.ContainerFile{
			Reader:            bytes.NewReader(jsonData),
			ContainerFilePath: "/tmp/tc-toxiproxy.json",
			FileMode:          0o644,
		})
		genericContainerReq.Cmd = append(genericContainerReq.Cmd, "-host=0.0.0.0", "-config=/tmp/tc-toxiproxy.json")
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{Container: container, proxiedEndpoints: make(map[int]string)}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	// Map the ports of the proxies to the container, so that we can use them in the tests
	for _, proxy := range settings.proxies {
		err := proxy.sanitize()
		if err != nil {
			return c, fmt.Errorf("sanitize: %w", err)
		}

		host, err := c.Host(ctx)
		if err != nil {
			return c, fmt.Errorf("host: %w", err)
		}

		mappedPort, err := c.MappedPort(ctx, nat.Port(fmt.Sprintf("%d/tcp", proxy.listenPort)))
		if err != nil {
			return c, fmt.Errorf("mapped port: %w", err)
		}

		c.proxiedEndpoints[proxy.listenPort] = net.JoinHostPort(host, mappedPort.Port())
	}

	return c, nil
}
