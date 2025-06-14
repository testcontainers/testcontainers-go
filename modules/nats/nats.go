package nats

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultClientPort     = "4222/tcp"
	defaultRoutingPort    = "6222/tcp"
	defaultMonitoringPort = "8222/tcp"
)

// NATSContainer represents the NATS container type used in the module
type NATSContainer struct {
	testcontainers.Container
	User     string
	Password string
}

// Deprecated: use Run instead
// RunContainer creates an instance of the NATS container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*NATSContainer, error) {
	return Run(ctx, "nats:2.9", opts...)
}

// Run creates an instance of the NATS container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*NATSContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultClientPort, defaultRoutingPort, defaultMonitoringPort),
		testcontainers.WithCmd("-DV", "-js"),
		testcontainers.WithWaitStrategy(wait.ForLog("Listening for client connections on 0.0.0.0:4222")),
	}

	moduleOpts = append(moduleOpts, opts...)

	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(CmdOption); ok {
			if err := apply(&settings); err != nil {
				return nil, fmt.Errorf("nats option: %w", err)
			}
		}
	}

	// Include the command line arguments
	cmdArgs := []string{}
	for k, v := range settings.CmdArgs {
		// always prepend the dash because it was removed in the options
		cmdArgs = append(cmdArgs, []string{"--" + k, v}...)
	}
	if len(cmdArgs) > 0 {
		moduleOpts = append(moduleOpts, testcontainers.WithCmdArgs(cmdArgs...))
	}

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *NATSContainer
	if container != nil {
		c = &NATSContainer{
			Container: container,
			User:      settings.CmdArgs["user"],
			Password:  settings.CmdArgs["pass"],
		}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}

func (c *NATSContainer) MustConnectionString(ctx context.Context) string {
	addr, err := c.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}
	return addr
}

// ConnectionString returns a connection string for the NATS container
func (c *NATSContainer) ConnectionString(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, defaultClientPort)
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	uri := fmt.Sprintf("nats://%s:%s", hostIP, mappedPort.Port())
	return uri, nil
}
