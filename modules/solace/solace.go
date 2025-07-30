package solace

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed mounts/solace.script.tpl
var customScriptTpl string

// Container represents a Solace container with additional settings
type Container struct {
	testcontainers.Container
	settings options
}

// Run starts a Solace container with the provided image and options
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// Override default options with provided ones
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, err
			}
		}
	}

	// Build wait strategies
	waitStrategies := make([]wait.Strategy, len(settings.services)+1)
	exposedPorts := make([]string, len(settings.services))

	// Primary wait strategy for Solace to be ready
	waitStrategies[0] = wait.ForExec([]string{"grep", "-q", "Primary Virtual Router is now active", "/usr/sw/jail/logs/system.log"}).
		WithStartupTimeout(1 * time.Minute).
		WithPollInterval(1 * time.Second)

	// Add port-based wait strategies for each service
	for i, service := range settings.services {
		port := fmt.Sprintf("%d/tcp", service.Port)
		waitStrategies[i+1] = wait.ForListeningPort(nat.Port(port))
		exposedPorts[i] = fmt.Sprintf("%d/tcp", service.Port)
	}

	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(exposedPorts...),
		testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.ShmSize = settings.shmSize
		}),
		testcontainers.WithWaitStrategy(wait.ForAll(waitStrategies...)),
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

	// Render CLI script for queue/topic configuration
	solaceScript, err := renderSolaceScript(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CLI script: %w", err)
	}

	// Copy the CLI script directly to the container
	err = c.CopyToContainer(ctx, []byte(solaceScript), "/usr/sw/jail/cliscripts/script.cli", 0o644)
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
func (c *Container) BrokerURLFor(ctx context.Context, service Service) (string, error) {
	p := nat.Port(fmt.Sprintf("%d/tcp", service.Port))
	return c.PortEndpoint(ctx, p, service.Protocol)
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

// renderSolaceScript renders the Solace CLI configuration script based on the provided settings.
// Reference: https://docs.solace.com/Admin-Ref/CLI-Reference/VMR_CLI_Commands.html
func renderSolaceScript(opts options) (string, error) {
	tmpl, err := template.New("solace-script").Parse(customScriptTpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse Solace script template: %w", err)
	}

	// Create template data structure
	data := struct {
		VPN      string
		Username string
		Password string
		Queues   map[string][]string
	}{
		VPN:      opts.vpn,
		Username: opts.username,
		Password: opts.password,
		Queues:   opts.queues,
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute Solace script template: %w", err)
	}

	return result.String(), nil
}
