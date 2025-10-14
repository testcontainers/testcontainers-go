package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort = "8200"
)

// VaultContainer represents the vault container type used in the module
type VaultContainer struct {
	testcontainers.Container
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Vault container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*VaultContainer, error) {
	return Run(ctx, "hashicorp/vault:1.13.0", opts...)
}

// Run creates an instance of the Vault container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*VaultContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultPort + "/tcp"),
		testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.CapAdd = []string{"CAP_IPC_LOCK"}
		}),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/v1/sys/health").WithPort(defaultPort)),
		testcontainers.WithEnv(map[string]string{
			"VAULT_ADDR": "http://0.0.0.0:" + defaultPort,
		}),
	}

	ctr, err := testcontainers.Run(ctx, img, append(moduleOpts, opts...)...)
	var c *VaultContainer
	if ctr != nil {
		c = &VaultContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run vault: %w", err)
	}

	return c, nil
}

// WithToken is a container option function that sets the root token for the Vault
func WithToken(token string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		return testcontainers.WithEnv(map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID": token,
			"VAULT_TOKEN":             token,
		})(req)
	}
}

// WithInitCommand is an option function that adds a set of initialization commands to the Vault's configuration
func WithInitCommand(commands ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		commandsList := make([]string, 0, len(commands))
		for _, command := range commands {
			commandsList = append(commandsList, "vault "+command)
		}
		cmd := []string{"/bin/sh", "-c", strings.Join(commandsList, " && ")}

		return testcontainers.WithAdditionalWaitStrategy(wait.ForExec(cmd))(req)
	}
}

// HttpHostAddress returns the http host address of Vault.
// It returns a string with the format http://<host>:<port>
//
//nolint:revive,staticcheck //FIXME
func (v *VaultContainer) HttpHostAddress(ctx context.Context) (string, error) {
	return v.PortEndpoint(ctx, defaultPort, "http")
}
