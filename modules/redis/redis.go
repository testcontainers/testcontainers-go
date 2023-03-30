package redis

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// defaultImage is the default image used for the redis container
const defaultImage = "docker.io/redis:7"

// redisServerProcess is the name of the redis server process
const redisServerProcess = "redis-server"

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

type RedisContainer struct {
	testcontainers.Container
}

func (c *RedisContainer) ConnectionString(ctx context.Context) (string, error) {
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

// StartContainer creates an instance of the Redis container type
func StartContainer(ctx context.Context, opts ...RedisContainerOption) (*RedisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultImage,
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("* Ready to accept connections"),
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

	return &RedisContainer{Container: container}, nil
}

// RedisContainerOption is a function that configures the redis container, affecting the container request
type RedisContainerOption func(req *testcontainers.ContainerRequest)

// WithImage sets the image to be used for the redis container
func WithImage(image string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Image = image
	}
}

// WithConfigFile sets the config file to be used for the redis container, and sets the command to run the redis server
// using the passed config file
func WithConfigFile(configFile string) func(req *testcontainers.ContainerRequest) {
	const defaultConfigFile = "/usr/local/redis.conf"

	return func(req *testcontainers.ContainerRequest) {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: defaultConfigFile,
			FileMode:          0755,
		}
		req.Files = append(req.Files, cf)

		if len(req.Cmd) == 0 {
			req.Cmd = []string{redisServerProcess, defaultConfigFile}
			return
		}

		// prepend the command to run the redis server with the config file, which must be the first argument of the redis server process
		if req.Cmd[0] == redisServerProcess {
			// just insert the config file, then the rest of the args
			req.Cmd = append([]string{redisServerProcess, defaultConfigFile}, req.Cmd[1:]...)
		} else if req.Cmd[0] != redisServerProcess {
			// prepend the redis server and the confif file, then the rest of the args
			req.Cmd = append([]string{redisServerProcess, defaultConfigFile}, req.Cmd...)
		}
	}
}

// WithLogLevel sets the log level for the redis server process
// See https://redis.io/docs/reference/modules/modules-api-ref/#redismodule_log for more information.
func WithLogLevel(level LogLevel) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		processRedisServerArgs(req, []string{"--loglevel", string(level)})
	}
}

// WithSnapshotting sets the snapshotting configuration for the redis server process. You can configure Redis to have it
// save the dataset every N seconds if there are at least M changes in the dataset.
// This method allows Redis to benefit from copy-on-write semantics.
// See https://redis.io/docs/management/persistence/#snapshotting for more information.
func WithSnapshotting(seconds int, changedKeys int) func(req *testcontainers.ContainerRequest) {
	if changedKeys < 1 {
		changedKeys = 1
	}
	if seconds < 1 {
		seconds = 1
	}

	return func(req *testcontainers.ContainerRequest) {
		processRedisServerArgs(req, []string{"--save", fmt.Sprintf("%d", seconds), fmt.Sprintf("%d", changedKeys)})
	}
}

func processRedisServerArgs(req *testcontainers.ContainerRequest, args []string) {
	if len(req.Cmd) == 0 {
		req.Cmd = append([]string{redisServerProcess}, args...)
		return
	}

	// prepend the command to run the redis server with the config file
	if req.Cmd[0] == redisServerProcess {
		// redis server is already set as the first argument, so just append the config file
		req.Cmd = append(req.Cmd, args...)
	} else if req.Cmd[0] != redisServerProcess {
		// redis server is not set as the first argument, so prepend it alongside the config file
		req.Cmd = append([]string{redisServerProcess}, req.Cmd...)
		req.Cmd = append(req.Cmd, args...)
	}
}
