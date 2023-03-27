package redis

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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
		Image:        "redis:6",
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
			req.Cmd = []string{"redis-server", defaultConfigFile}
			return
		}

		// prepend the command to run the redis server with the config file
		if req.Cmd[0] == "redis-server" {
			// redis server is already set as the first argument, so just append the config file
			req.Cmd = append([]string{"redis-server", defaultConfigFile}, req.Cmd[1:]...)
		} else if req.Cmd[0] != "redis-server" {
			// redis server is not set as the first argument, so prepend it alongside the config file
			req.Cmd = append([]string{"redis-server", defaultConfigFile}, req.Cmd...)
		}
	}
}
