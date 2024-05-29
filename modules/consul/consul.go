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

// Container represents the Consul container type used in the module.
type Container struct {
	*testcontainers.DockerContainer
}

// ApiEndpoint returns host:port for the HTTP API endpoint.
func (c *Container) ApiEndpoint(ctx context.Context) (string, error) {
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
	return func(req *testcontainers.Request) error {
		req.Env["CONSUL_LOCAL_CONFIG"] = config

		return nil
	}
}

// WithConfigFile takes in a path to a JSON file to define a configuration to be used by the instance.
func WithConfigFile(configPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.Request) error {
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
func RunContainer(ctx context.Context, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	req := testcontainers.Request{
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
		Started: true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.New(ctx, req)
	if err != nil {
		return nil, err
	}

	return &Container{DockerContainer: container}, nil
}
