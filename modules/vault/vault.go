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
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{defaultPort + "/tcp"},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.CapAdd = []string{"IPC_LOCK"}
		},
		WaitingFor: wait.ForHTTP("/v1/sys/health").WithPort(defaultPort),
		Env: map[string]string{
			"VAULT_ADDR": "http://0.0.0.0:" + defaultPort,
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *VaultContainer
	if container != nil {
		c = &VaultContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// WithToken is a container option function that sets the root token for the Vault
func WithToken(token string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["VAULT_DEV_ROOT_TOKEN_ID"] = token
		req.Env["VAULT_TOKEN"] = token

		return nil
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

		req.WaitingFor = wait.ForAll(req.WaitingFor, wait.ForExec(cmd))

		return nil
	}
}

// HttpHostAddress returns the http host address of Vault.
// It returns a string with the format http://<host>:<port>
func (v *VaultContainer) HttpHostAddress(ctx context.Context) (string, error) {
	host, err := v.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := v.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%d", host, port.Int()), nil
}
