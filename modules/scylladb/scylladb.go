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
	port           = "9042/tcp"
	shardAwarePort = "19042/tcp"
)

// Container represents a ScyllaDB container type used in the module
type Container struct {
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
			FileMode:          0o644,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

// WithShardAwareness enable shard-awareness in the ScyllaDB container so you can use the `19042` port.
func WithShardAwareness() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.ExposedPorts = append(req.ExposedPorts, shardAwarePort)
		req.WaitingFor = wait.ForAll(req.WaitingFor, wait.ForListeningPort(shardAwarePort))
		return nil
	}
}

// WithAlternator enables the Alternator (DynamoDB Compatible API) service in the ScyllaDB container.
// It will set the "alternator-port" parameter to the specified port.
// It will also set the "alternator-write-isolation" parameter to "always" as a command line argument to the container.
func WithAlternator(alternatorPort uint16) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		alternatorPortStr := strconv.FormatInt(int64(alternatorPort), 10)
		req.ExposedPorts = append(req.ExposedPorts, alternatorPortStr)
		req.WaitingFor = wait.ForAll(req.WaitingFor, wait.ForListeningPort(nat.Port(alternatorPortStr)))
		setCommandFlag(req, map[string]string{
			"--alternator-port":            alternatorPortStr,
			"--alternator-write-isolation": "always",
		})

		return nil
	}
}

// WithCustomCommands sets custom commands with  values for the ScyllaDB container.
// This is an option to overwrite the default commands with a custom one.
// See more [here](https://opensource.docs.scylladb.com/stable/operating-scylla/procedures/tips/best-practices-scylla-on-docker.html)
func WithCustomCommands(cmds map[string]string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		setCommandFlag(req, cmds)
		return nil
	}
}

// ConnectionHost returns the host and port of the Scylladb container with the default port
// and obtaining the host and exposed port from the container
// Eg: "host:port" -> "localhost:9042" -> "localhost:19042" -> "localhost:8000"
func (c Container) ConnectionHost(ctx context.Context, port int) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	containerPort, err := c.MappedPort(ctx, nat.Port(strconv.Itoa(port)))
	if err != nil {
		return "", err
	}

	return host + ":" + containerPort.Port(), nil
}

// Run starts a ScyllaDB container with the specified image and options
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{port},
		Cmd: []string{
			"--developer-mode=1",
			"--overprovisioned=1",
			"--smp=1",
			"--memory=512M",
		},
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
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

func setCommandFlag(req *testcontainers.GenericContainerRequest, flag map[string]string) {
	cmds := make([]string, 0, len(req.Cmd)+len(flag))

	for _, cmd := range req.Cmd {
		seps := strings.SplitN(cmd, "=", 1)
		val, ok := flag[seps[0]]

		if !ok {
			cmds = append(cmds, cmd)
		} else {
			cmds = append(cmds, fmt.Sprintf("%s=%s", seps[0], val))
			delete(flag, seps[0])
		}
	}

	for key, val := range flag {
		cmds = append(cmds, fmt.Sprintf("%s=%s", key, val))
	}

	req.Cmd = cmds
}
