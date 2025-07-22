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

// waitForSolaceActive waits for the Solace broker to be fully active by checking the system log
func (s *SolaceContainer) waitForSolaceActive(ctx context.Context) error {
	const maxAttempts = 60
	const intervalSeconds = 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Execute grep command to check for "Primary Virtual Router is now active" in system log
		code, _, err := s.Exec(ctx, []string{"grep", "-R", "Primary Virtual Router is now active", "/usr/sw/jail/logs/system.log"})

		if err == nil && code == 0 {
			// Found the message, broker is active
			return nil
		}

		// Wait before next attempt
		time.Sleep(intervalSeconds * time.Second)
	}

	return fmt.Errorf("timeout waiting for Solace broker to become active after %d seconds", maxAttempts)
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

	// --- CLI script generation for queue/topic config ---
	// Reference: https://docs.solace.com/Admin-Ref/CLI-Reference/VMR_CLI_Commands.html
	var cliScript string
	if len(settings.queues) > 0 {
		cliScript += "enable\nconfigure\n"

		// Create VPN if not default
		if settings.vpn != defaultVpn {
			cliScript += fmt.Sprintf("create message-vpn %s\n", settings.vpn)
			cliScript += "no shutdown\n"
			cliScript += "exit\n"
			cliScript += fmt.Sprintf("client-profile default message-vpn %s\n", settings.vpn)
			cliScript += "message-spool\n"
			cliScript += "allow-guaranteed-message-send\n"
			cliScript += "allow-guaranteed-message-receive\n"
			cliScript += "allow-guaranteed-endpoint-create\n"
			cliScript += "allow-guaranteed-endpoint-create-durability all\n"
			cliScript += "exit\n"
			cliScript += "exit\n"
			cliScript += fmt.Sprintf("message-spool message-vpn %s\n", settings.vpn)
			cliScript += "max-spool-usage 60000\n"
			cliScript += "exit\n"
		}

		// Configure username and password
		cliScript += fmt.Sprintf("create client-username %s message-vpn %s\n", settings.username, settings.vpn)
		cliScript += fmt.Sprintf("password %s\n", settings.password)
		cliScript += "no shutdown\n"
		cliScript += "exit\n"

		// Configure VPN Basic authentication
		cliScript += fmt.Sprintf("message-vpn %s\n", settings.vpn)
		cliScript += "authentication basic auth-type internal\n"
		cliScript += "no shutdown\n"
		cliScript += "end\n"

		// Configure queues and topic subscriptions
		cliScript += "configure\n"
		cliScript += fmt.Sprintf("message-spool message-vpn %s\n", settings.vpn)
		for queue, topics := range settings.queues {
			// Create the queue first
			cliScript += fmt.Sprintf("create queue %s\n", queue)
			cliScript += "access-type exclusive\n"
			cliScript += "permission all consume\n"
			cliScript += "no shutdown\n"
			cliScript += "exit\n"

			// Add topic subscriptions to the queue
			for _, topic := range topics {
				cliScript += fmt.Sprintf("queue %s\n", queue)
				cliScript += fmt.Sprintf("subscription topic %s\n", topic)
				cliScript += "exit\n"
			}
		}
		cliScript += "exit\nexit\n"
	}

	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: settings.exposedPorts,
		Env:          settings.envVars,
		Cmd:          nil,
		WaitingFor:   wait.ForHTTP("/").WithPort("8080/tcp").WithStartupTimeout(1 * time.Minute),
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

	// Wait for Solace broker to be fully active before configuring
	if err := c.waitForSolaceActive(ctx); err != nil {
		return nil, fmt.Errorf("failed waiting for Solace broker to become active: %w", err)
	}

	// Copy and execute CLI script inside the container if it was generated
	if cliScript != "" {
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
