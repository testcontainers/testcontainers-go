package solace

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultVpn = "default"
)

type SolaceContainer struct {
	testcontainers.Container
	settings options
}

func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*SolaceContainer, error) {
	// Default to the standard Solace image if none provided
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, err
			}
		}
	}

	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: settings.exposedPorts,
		Env:          settings.envVars,
		Cmd:          nil,
		WaitingFor: wait.ForExec([]string{"grep", "-q", "Primary Virtual Router is now active", "/usr/sw/jail/logs/system.log"}).
			WithStartupTimeout(1 * time.Minute).
			WithPollInterval(1 * time.Second),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.ShmSize = settings.shmSize
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	var c *SolaceContainer
	if container != nil {
		c = &SolaceContainer{
			Container: container,
			settings:  settings,
		}
	}

	// Generate CLI script for queue/topic configuration
	cliScript := generateCLIScript(settings)
	if cliScript == "" {
		return c, nil
	}
	// Write the CLI script to a temp file
	tmpFile, err := os.CreateTemp("", "solace-queue-setup-*.cli")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp CLI script: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(cliScript)); err != nil {
		return nil, fmt.Errorf("failed to write CLI script: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close CLI script: %w", err)
	}

	// Copy the script into the container at the correct location
	err = c.CopyFileToContainer(ctx, tmpFile.Name(), "/usr/sw/jail/cliscripts/script.cli", 0o644)
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
func (s *SolaceContainer) BrokerURLFor(ctx context.Context, service Service) (string, error) {
	p := nat.Port(fmt.Sprintf("%d/tcp", service.Port))
	return s.PortEndpoint(ctx, p, service.Protocol)
}

func (c *SolaceContainer) Username() string {
	return c.settings.username
}

func (c *SolaceContainer) Password() string {
	return c.settings.password
}

func (c *SolaceContainer) Vpn() string {
	return c.settings.vpn
}
