package vault

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// vaultContainer represents the vault container type used in the module
type vaultContainer struct {
	testcontainers.Container
	config *Config
}

// StartContainer creates an instance of the vault container type
func StartContainer(ctx context.Context, opts ...Option) (*vaultContainer, error) {
	config := &Config{
		imageName: "vault:1.13.0",
		port:      8200,
		secrets:   map[string][]string{},
	}

	for _, opt := range opts {
		opt(config)
	}

	req := testcontainers.ContainerRequest{
		Image:        config.imageName,
		ExposedPorts: []string{fmt.Sprintf("%d/tcp", config.port)},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.CapAdd = []string{"IPC_LOCK"}
		},
		WaitingFor: wait.ForHTTP("/v1/sys/health").WithPort(nat.Port(strconv.Itoa(config.port))),
		Env:        config.exportEnv(),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	v := vaultContainer{container, config}

	if err = v.addSecrets(ctx); err != nil {
		return nil, err
	}

	if err = v.runInitCommands(ctx); err != nil {
		return nil, err
	}

	return &v, nil
}

func (v *vaultContainer) HttpHostAddress(ctx context.Context) (string, error) {
	host, err := v.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := v.MappedPort(ctx, nat.Port(strconv.Itoa(v.config.port)))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%d", host, port.Int()), nil
}

func (v *vaultContainer) addSecrets(ctx context.Context) error {
	if len(v.config.secrets) == 0 {
		return nil
	}

	code, _, err := v.Exec(ctx, buildExecCommand(v.config.secrets))
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("failed to add secrets %v to Vault via exec command: %d", v.config.secrets, code)
	}

	return nil
}

func (v *vaultContainer) runInitCommands(ctx context.Context) error {
	if len(v.config.initCommands) == 0 {
		return nil
	}

	commands := make([]string, 0, len(v.config.initCommands))
	for _, command := range v.config.initCommands {
		commands = append(commands, "vault "+command)
	}
	fullCommand := []string{"/bin/sh", "-c", strings.Join(commands, " && ")}

	code, _, err := v.Exec(ctx, fullCommand)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("failed to execute init commands: exit code %d", code)
	}

	return nil
}

func buildExecCommand(secretsMap map[string][]string) []string {
	var commandParts []string

	// Loop over the secrets map and build the command string
	for path, secrets := range secretsMap {
		commandParts = append(commandParts, "vault", "kv", "put", path)
		commandParts = append(commandParts, secrets...)
		commandParts = append(commandParts, "&&")
	}

	if len(commandParts) > 0 {
		commandParts = commandParts[:len(commandParts)-1]
	}

	// Prepend the command with "/bin/sh -c" to execute the command in a shell
	commandParts = append([]string{"/bin/sh", "-c"}, strings.Join(commandParts, " "))

	return commandParts
}
