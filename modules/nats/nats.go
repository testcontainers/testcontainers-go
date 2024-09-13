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

// Container represents the NATS container type used in the module
type Container struct {
	*testcontainers.DockerContainer
	User     string
	Password string
}

// Run creates an instance of the NATS container type
func Run(ctx context.Context, img string, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	req := testcontainers.Request{
		Image:        img,
		ExposedPorts: []string{defaultClientPort, defaultRoutingPort, defaultMonitoringPort},
		Cmd:          []string{"-DV", "-js"},
		WaitingFor:   wait.ForLog("Listening for client connections on 0.0.0.0:4222"),
		Started:      true,
	}

	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(CmdOption); ok {
			apply(&settings)
		}
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	// Include the command line arguments
	for k, v := range settings.CmdArgs {
		// always prepend the dash because it was removed in the options
		req.Cmd = append(req.Cmd, []string{"--" + k, v}...)
	}

	ctr, err := testcontainers.Run(ctx, req)
	var c *Container
	if ctr != nil {
		c = &Container{
			DockerContainer: ctr,
			User:            settings.CmdArgs["user"],
			Password:        settings.CmdArgs["pass"],
		}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

func (c *Container) MustConnectionString(ctx context.Context, args ...string) string {
	addr, err := c.ConnectionString(ctx, args...)
	if err != nil {
		panic(err)
	}
	return addr
}

// ConnectionString returns a connection string for the NATS container
func (c *Container) ConnectionString(ctx context.Context, args ...string) (string, error) {
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
