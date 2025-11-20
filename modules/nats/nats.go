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
	return Run(ctx, "nats:2.11.7", opts...)
}

// Run creates an instance of the NATS container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*NATSContainer, error) {
	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(CmdOption); ok {
			apply(&settings)
		}
	}

	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultClientPort, defaultRoutingPort, defaultMonitoringPort),
		testcontainers.WithCmd("-DV", "-js"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(defaultClientPort)),
	}

	moduleOpts = append(moduleOpts, opts...)

	// Include the command line arguments
	cmdArgs := []string{}
	for k, v := range settings.CmdArgs {
		// always prepend the dash because it was removed in the options
		cmdArgs = append(cmdArgs, "--"+k, v)
	}
	if len(cmdArgs) > 0 {
		moduleOpts = append(moduleOpts, testcontainers.WithCmdArgs(cmdArgs...))
	}

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *NATSContainer
	if ctr != nil {
		c = &NATSContainer{
			Container: ctr,
			User:      settings.CmdArgs["user"],
			Password:  settings.CmdArgs["pass"],
		}
	}

	if err != nil {
		return c, fmt.Errorf("run nats: %w", err)
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
	return c.PortEndpoint(ctx, defaultClientPort, "nats")
}
