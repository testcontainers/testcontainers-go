package consul

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultHTTPAPIPort = "8500"
	defaultBrokerPort  = "8600"
)

const (
	// Deprecated: it will be removed in the next major version.
	DefaultBaseImage = "hashicorp/consul:1.15"
)

// ConsulContainer represents the Consul container type used in the module.
type ConsulContainer struct {
	testcontainers.Container
}

// ApiEndpoint returns host:port for the HTTP API endpoint.
//
//nolint:revive,staticcheck //FIXME
func (c *ConsulContainer) ApiEndpoint(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, defaultHTTPAPIPort)
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	return uri, nil
}

// WithConfigString takes in a JSON string of keys and values to define a configuration to be used by the instance.
func WithConfigString(config string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["CONSUL_LOCAL_CONFIG"] = config

		return nil
	}
}

// WithConfigFile takes in a path to a JSON file to define a configuration to be used by the instance.
func WithConfigFile(configPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configPath,
			ContainerFilePath: "/consul/config/node.json",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Consul container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ConsulContainer, error) {
	return Run(ctx, "hashicorp/consul:1.15", opts...)
}

// Run creates an instance of the Consul container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ConsulContainer, error) {
	containerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: img,
			ExposedPorts: []string{
				defaultHTTPAPIPort + "/tcp",
				defaultBrokerPort + "/tcp",
				defaultBrokerPort + "/udp",
			},
			Env: map[string]string{},
			WaitingFor: wait.ForAll(
				wait.ForLog("Consul agent running!"),
				wait.ForListeningPort(defaultHTTPAPIPort+"/tcp"),
			),
		},
		Started: true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&containerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, containerReq)
	var c *ConsulContainer
	if container != nil {
		c = &ConsulContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
