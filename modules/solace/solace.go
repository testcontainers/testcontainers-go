package solace

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultVPN = "default"
)

// Container represents a Solace container with additional settings
type Container struct {
	testcontainers.Container
	settings options
}

// Run starts a Solace container with the provided image and options
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// Default to the standard Solace image if none provided
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, err
			}
		}
	}

	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(settings.exposedPorts...),
		testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.ShmSize = settings.shmSize
		}),
		testcontainers.WithWaitStrategy(wait.ForExec([]string{"grep", "-q", "Primary Virtual Router is now active", "/usr/sw/jail/logs/system.log"}).
			WithStartupTimeout(1 * time.Minute).
			WithPollInterval(1 * time.Second)),
	}

	moduleOpts = append(moduleOpts, opts...)
	container, err := testcontainers.Run(ctx, img, moduleOpts...)

	var c *Container
	if container != nil {
		c = &Container{
			Container: container,
			settings:  settings,
		}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	// Generate CLI script for queue/topic configuration
	cliScript := generateCLIScript(settings)
	if cliScript == "" {
		return c, nil
	}

	// Copy the CLI script directly to the container
	err = c.CopyToContainer(ctx, []byte(cliScript), "/usr/sw/jail/cliscripts/script.cli", 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to copy CLI script to container: %w", err)
	}

	// Execute the script
	code, out, err := c.Exec(ctx, []string{"/usr/sw/loads/currentload/bin/cli", "-A", "-es", "script.cli"})
	output := ""
	if out != nil {
		bytes, readErr := io.ReadAll(out)
		if readErr == nil {
			output = string(bytes)
		} else {
			output = fmt.Sprintf("[ERROR reading CLI output: %v]", readErr)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute CLI script for queue/topic setup: %w", err)
	}
	if code != 0 {
		return nil, fmt.Errorf("CLI script execution failed with exit code %d: %s", code, output)
	}

	return c, nil
}

// BrokerURLFor returns the origin URL for a given service
func (s *Container) BrokerURLFor(ctx context.Context, service Service) (string, error) {
	p := nat.Port(fmt.Sprintf("%d/tcp", service.Port))
	return s.PortEndpoint(ctx, p, service.Protocol)
}

// Username returns the username configured for the Solace container
func (c *Container) Username() string {
	return c.settings.username
}

// Password returns the password configured for the Solace container
func (c *Container) Password() string {
	return c.settings.password
}

// Vpn returns the VPN name configured for the Solace container
func (c *Container) VPN() string {
	return c.settings.vpn
}
