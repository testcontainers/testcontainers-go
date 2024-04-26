package consul

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultHttpApiPort = "8500"
	defaultBrokerPort  = "8600"
)

const (
	DefaultBaseImage = "docker.io/hashicorp/consul:1.15"
)

// ConsulContainer represents the Consul container type used in the module.
type ConsulContainer struct {
	testcontainers.Container
}

// ApiEndpoint returns host:port for the HTTP API endpoint.
func (c *ConsulContainer) ApiEndpoint(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, defaultHttpApiPort)
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

// RunContainer creates an instance of the Consul container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ConsulContainer, error) {
	containerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: DefaultBaseImage,
			ExposedPorts: []string{
				defaultHttpApiPort + "/tcp",
				defaultBrokerPort + "/tcp",
				defaultBrokerPort + "/udp",
			},
			Env: map[string]string{},
			WaitingFor: wait.ForAll(
				wait.ForLog("Consul agent running!"),
				wait.ForListeningPort(defaultHttpApiPort+"/tcp"),
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
	if err != nil {
		return nil, err
	}

	return &ConsulContainer{Container: container}, nil
}
