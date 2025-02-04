package valkey

import (
	"context"
	"fmt"
	"strconv"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ValkeyContainer represents the Valkey container type used in the module
type ValkeyContainer struct {
	testcontainers.Container
}

// valkeyServerProcess is the name of the valkey server process
const valkeyServerProcess = "valkey-server"

type LogLevel string

const (
	// LogLevelDebug is the debug log level
	LogLevelDebug LogLevel = "debug"
	// LogLevelVerbose is the verbose log level
	LogLevelVerbose LogLevel = "verbose"
	// LogLevelNotice is the notice log level
	LogLevelNotice LogLevel = "notice"
	// LogLevelWarning is the warning log level
	LogLevelWarning LogLevel = "warning"
)

// ConnectionString returns the connection string for the Valkey container
func (c *ValkeyContainer) ConnectionString(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, "6379/tcp")
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	uri := fmt.Sprintf("redis://%s:%s", hostIP, mappedPort.Port())
	return uri, nil
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Valkey container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ValkeyContainer, error) {
	return Run(ctx, "valkey/valkey:7.2.5", opts...)
}

// Run creates an instance of the Valkey container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ValkeyContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
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
	var c *ValkeyContainer
	if container != nil {
		c = &ValkeyContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// WithConfigFile sets the config file to be used for the valkey container, and sets the command to run the valkey server
// using the passed config file
func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	const defaultConfigFile = "/usr/local/valkey.conf"

	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: defaultConfigFile,
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		if len(req.Cmd) == 0 {
			req.Cmd = []string{valkeyServerProcess, defaultConfigFile}
			return nil
		}

		// prepend the command to run the redis server with the config file, which must be the first argument of the redis server process
		if req.Cmd[0] == valkeyServerProcess {
			// just insert the config file, then the rest of the args
			req.Cmd = append([]string{valkeyServerProcess, defaultConfigFile}, req.Cmd[1:]...)
		} else if req.Cmd[0] != valkeyServerProcess {
			// prepend the redis server and the config file, then the rest of the args
			req.Cmd = append([]string{valkeyServerProcess, defaultConfigFile}, req.Cmd...)
		}

		return nil
	}
}

// WithLogLevel sets the log level for the valkey server process
// See https://redis.io/docs/reference/modules/modules-api-ref/#redismodule_log for more information.
func WithLogLevel(level LogLevel) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		processValkeyServerArgs(req, []string{"--loglevel", string(level)})

		return nil
	}
}

// WithSnapshotting sets the snapshotting configuration for the valkey server process. You can configure Valkey to have it
// save the dataset every N seconds if there are at least M changes in the dataset.
// This method allows Valkey to benefit from copy-on-write semantics.
// See https://redis.io/docs/management/persistence/#snapshotting for more information.
func WithSnapshotting(seconds int, changedKeys int) testcontainers.CustomizeRequestOption {
	if changedKeys < 1 {
		changedKeys = 1
	}
	if seconds < 1 {
		seconds = 1
	}

	return func(req *testcontainers.GenericContainerRequest) error {
		processValkeyServerArgs(req, []string{"--save", strconv.Itoa(seconds), strconv.Itoa(changedKeys)})
		return nil
	}
}

func processValkeyServerArgs(req *testcontainers.GenericContainerRequest, args []string) {
	if len(req.Cmd) == 0 {
		req.Cmd = append([]string{valkeyServerProcess}, args...)
		return
	}

	// prepend the command to run the valkey server with the config file
	if req.Cmd[0] == valkeyServerProcess {
		// valkey server is already set as the first argument, so just append the config file
		req.Cmd = append(req.Cmd, args...)
	} else if req.Cmd[0] != valkeyServerProcess {
		// valkey server is not set as the first argument, so prepend it alongside the config file
		req.Cmd = append([]string{valkeyServerProcess}, req.Cmd...)
		req.Cmd = append(req.Cmd, args...)
	}
}
