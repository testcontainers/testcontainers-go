package dockermcpgateway

import (
	"errors"
	"maps"
	"slices"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	tools   map[string][]string
	secrets map[string]string
}

func defaultOptions() options {
	return options{
		tools:   map[string][]string{},
		secrets: map[string]string{},
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the DockerMCPGateway container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithTools sets a server's tools to use in the DockerMCPGateway container.
// Multiple calls to this function with the same server will append to the existing tools for that server.
// No duplicate tools will be added for the same server.
func WithTools(server string, tools []string) Option {
	return func(o *options) error {
		if server == "" {
			return errors.New("server cannot be empty")
		}
		if len(tools) == 0 {
			return errors.New("tools cannot be empty")
		}

		if slices.Contains(tools, "") {
			return errors.New("tool cannot be empty")
		}

		currentTools, exists := o.tools[server]
		if exists {
			// Append only unique tools to avoid duplicates
			for _, tool := range tools {
				if !slices.Contains(currentTools, tool) {
					currentTools = append(currentTools, tool)
				}
			}
			o.tools[server] = currentTools
		} else {
			// If the server does not exist, create a new entry
			o.tools[server] = tools
		}

		return nil
	}
}

// WithServers sets the servers to use in the DockerMCPGateway container.
// Multiple calls to this function will append to the existing values.
func WithSecret(key, value string) Option {
	return func(o *options) error {
		if key == "" {
			return errors.New("secret key cannot be empty")
		}

		o.secrets[key] = value
		return nil
	}
}

// WithSecrets sets the secrets to use in the DockerMCPGateway container.
// Multiple calls to this function will merge the secrets into the existing map.
func WithSecrets(secrets map[string]string) Option {
	return func(o *options) error {
		if len(secrets) == 0 {
			return errors.New("secrets cannot be empty")
		}

		maps.Copy(o.secrets, secrets)
		return nil
	}
}
