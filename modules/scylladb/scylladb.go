package scylladb

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	port           = nat.Port("9042/tcp")
	shardAwarePort = nat.Port("19042/tcp")
)

type ScyllaDBContainer struct {
	testcontainers.Container
}

// WithConfigFile sets the YAML config file to be used for the cassandra container
// It will also set the "configFile" parameter to the path of the config file
// as a command line argument to the container.
func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/scylla/scylla.yaml",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

// WithShardAwareness enable shard-awareness in the ScyllaDB container so you can use the `19042` port.
func WithShardAwareness() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.ExposedPorts = append(req.ExposedPorts, shardAwarePort.Port())
		req.WaitingFor = wait.ForAll(req.WaitingFor, wait.ForListeningPort(shardAwarePort))
		return nil
	}
}

func WithoutDeveloperMode() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		setCommandFlag(req, "--developer-mode", "0")
		return nil
	}
}

// WithMemoryLimit sets the memory limit for the ScyllaDB container
// Eg: "1G" | "2G" | "512M"
func WithMemoryLimit(memoryLimit string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		setCommandFlag(req, "--memory", memoryLimit)
		return nil
	}
}

// WithCpuLimit sets the CPU/Cores limit for the ScyllaDB container
// If you're using Shard Awareness, you should set the CPU limit to at least 2 cores.
func WithCpuLimit(cores uint) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		setCommandFlag(req, "--smp", strconv.Itoa(int(cores)))
		return nil
	}
}

// ShardAwareConnectionHost returns the host and port of the Scylladb container
// using the shard-aware port, and obtaining the host and exposed port from the container
// Eg: "host:port" -> "localhost:19042"
func (c ScyllaDBContainer) ShardAwareConnectionHost(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, shardAwarePort)
	if err != nil {
		return "", err
	}

	return host + ":" + port.Port(), nil
}

// ConnectionHost returns the host and port of the Scylladb container with the default port
// and obtaining the host and exposed port from the container
// Eg: "host:port" -> "localhost:9042"
func (c ScyllaDBContainer) ConnectionHost(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}

	return host + ":" + port.Port(), nil
}

// Run starts a ScyllaDB container with the specified image and options
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ScyllaDBContainer, error) {
	defaultCommands := []string{
		"--smp=1",
		"--memory=1G",
		"--developer-mode=1",
		"--overprovisioned=1",
	}

	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{port.Port()},
		Cmd:          defaultCommands,
		WaitingFor: wait.ForAll(
			wait.ForLog(".*initialization completed.").AsRegexp(),
			wait.ForListeningPort(port),
			wait.ForExec([]string{"cqlsh", "-e", "SELECT bootstrapped FROM system.local"}).WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return strings.Contains(string(data), "COMPLETED")
			}),
		),
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

	fmt.Println(genericContainerReq.Cmd)

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *ScyllaDBContainer
	if container != nil {
		c = &ScyllaDBContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

func setCommandFlag(req *testcontainers.GenericContainerRequest, flag, value string) {
	cmdsWithoutDeveloperMode := make([]string, 0, len(req.Cmd))
	for _, cmd := range req.Cmd {
		if !strings.Contains(cmd, flag) {
			cmdsWithoutDeveloperMode = append(cmdsWithoutDeveloperMode, cmd)
		}
	}
	req.Cmd = cmdsWithoutDeveloperMode
	req.Cmd = append(req.Cmd, fmt.Sprintf("%v=%v", flag, value))
}
