package dockermcpgateway

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort = "8811/tcp"
	secretsPath = "/testcontainers/app/secrets"
)

// Container represents the DockerMCPGateway container type used in the module
type Container struct {
	testcontainers.Container
	tools map[string][]string
}

// Run creates an instance of the DockerMCPGateway container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	dockerHostMount := core.MustExtractDockerSocket(ctx)

	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.Binds = []string{
				dockerHostMount + ":/var/run/docker.sock",
			}
		}),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort(defaultPort),
			wait.ForLog(".*Start sse server on port.*").AsRegexp(),
		)),
	}

	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, err
			}
		}
	}

	cmds := []string{"--transport=sse"}
	for server, tools := range settings.tools {
		cmds = append(cmds, "--servers="+server)
		for _, tool := range tools {
			cmds = append(cmds, "--tools="+tool)
		}
	}
	if len(settings.secrets) > 0 {
		cmds = append(cmds, "--secrets="+secretsPath)

		secretsContent := ""
		for key, value := range settings.secrets {
			secretsContent += key + "=" + value + "\n"
		}

		moduleOpts = append(moduleOpts, testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader:            strings.NewReader(secretsContent),
			ContainerFilePath: secretsPath,
			FileMode:          0o644,
		}))
	}

	moduleOpts = append(moduleOpts, testcontainers.WithCmd(cmds...))

	// append user-defined options
	moduleOpts = append(moduleOpts, opts...)

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if container != nil {
		c = &Container{Container: container, tools: settings.tools}
	}

	if err != nil {
		return c, fmt.Errorf("run mcp gateway: %w", err)
	}

	return c, nil
}

// GatewayEndpoint returns the endpoint for the DockerMCPGateway container.
// It uses the mapped port for the default port (8811/tcp) and the "http" protocol.
func (c *Container) GatewayEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, defaultPort, "http")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return endpoint, nil
}

// Tools returns the tools configured for the DockerMCPGateway container,
// indexed by server name.
// The keys are the server names and the values are slices of tool names.
func (c *Container) Tools() map[string][]string {
	return c.tools
}
