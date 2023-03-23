package vault

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go/wait"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/testcontainers/testcontainers-go"
)

const (
	defaultPort      = "8200"
	defaultImageName = "hashicorp/vault:1.13.0"
)

// ContainerOptions is a function that can be used to configure the Vault container
type ContainerOptions func(req *testcontainers.ContainerRequest)

// VaultContainer represents the vault container type used in the module
type VaultContainer struct {
	testcontainers.Container
}

// StartContainer creates an instance of the vault container type
func StartContainer(ctx context.Context, opts ...ContainerOptions) (*VaultContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultImageName,
		ExposedPorts: []string{defaultPort + "/tcp"},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.CapAdd = []string{"IPC_LOCK"}
		},
		WaitingFor: wait.ForHTTP("/v1/sys/health").WithPort(defaultPort),
		Env: map[string]string{
			"VAULT_ADDR": "http://0.0.0.0:" + defaultPort,
		},
	}

	for _, opt := range opts {
		opt(&req)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &VaultContainer{container}, nil
}

// WithImageName is an option function that sets the Docker image name for the Vault
func WithImageName(imageName string) ContainerOptions {
	return func(req *testcontainers.ContainerRequest) {
		req.Image = imageName
	}
}

// WithToken is a container option function that sets the root token for the Vault
func WithToken(token string) ContainerOptions {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["VAULT_DEV_ROOT_TOKEN_ID"] = token
		req.Env["VAULT_TOKEN"] = token
	}
}

// WithInitCommand is an option function that adds a set of initialization commands to the Vault's configuration
func WithInitCommand(commands ...string) ContainerOptions {
	return func(req *testcontainers.ContainerRequest) {
		commandsList := make([]string, 0, len(commands))
		for _, command := range commands {
			commandsList = append(commandsList, "vault "+command)
		}
		cmd := []string{"/bin/sh", "-c", strings.Join(commandsList, " && ")}

		req.WaitingFor = wait.ForAll(req.WaitingFor, wait.ForExec(cmd))
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
