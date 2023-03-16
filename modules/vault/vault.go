package vault

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"io"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	errorAddSecret = "failed to add secrets %v to Vault via exec command: %w"
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

func (v *vaultContainer) addSecrets(ctx context.Context) error {
	if len(v.config.secrets) == 0 {
		return nil
	}

	code, reader, err := v.Exec(ctx, buildExecCommand(v.config.secrets))
	if err != nil || code != 0 {
		return err
	}

	if code != 0 {
		return fmt.Errorf(errorAddSecret, v.config.secrets, err)
	}

	out, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf(errorAddSecret, v.config.secrets, err)
	}

	if !strings.Contains(string(out), "Success") {
		return fmt.Errorf(errorAddSecret, v.config.secrets, err)
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

	code, stdout, err := v.Exec(ctx, fullCommand)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("failed to execute init commands: exit code %d, stdout %s", code, stdout)
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
