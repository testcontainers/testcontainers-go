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
	return c.PortEndpoint(ctx, defaultHTTPAPIPort, "")
}

// WithConfigString takes in a JSON string of keys and values to define a configuration to be used by the instance.
func WithConfigString(config string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"CONSUL_LOCAL_CONFIG": config,
	})
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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultHTTPAPIPort+"/tcp", defaultBrokerPort+"/tcp", defaultBrokerPort+"/udp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("Consul agent running!"),
			wait.ForListeningPort(defaultHTTPAPIPort+"/tcp"),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *ConsulContainer
	if ctr != nil {
		c = &ConsulContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run consul: %w", err)
	}

	return c, nil
}
