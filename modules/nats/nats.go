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
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{defaultClientPort, defaultRoutingPort, defaultMonitoringPort},
		Cmd:          []string{"-DV", "-js"},
		WaitingFor: wait.ForAll(
			wait.ForExposedPortOnly(defaultClientPort),
			wait.ForExposedPortOnly(defaultRoutingPort),
			wait.ForExposedPortOnly(defaultMonitoringPort),
		),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(CmdOption); ok {
			apply(&settings)
		}
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	// Include the command line arguments
	for k, v := range settings.CmdArgs {
		// always prepend the dash because it was removed in the options
		genericContainerReq.Cmd = append(genericContainerReq.Cmd, []string{"--" + k, v}...)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	natsContainer := NATSContainer{
		Container: container,
		User:      settings.CmdArgs["user"],
		Password:  settings.CmdArgs["pass"],
	}

	return &natsContainer, nil
}

func (c *NATSContainer) MustConnectionString(ctx context.Context, args ...string) string {
	addr, err := c.ConnectionString(ctx, args...)
	if err != nil {
		panic(err)
	}
	return addr
}

// ConnectionString returns a connection string for the NATS container
func (c *NATSContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
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
