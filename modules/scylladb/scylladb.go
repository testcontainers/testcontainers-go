package scylladb

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	port           = "9042/tcp"
	shardAwarePort = "19042/tcp"
	alternatorPort = "8000/tcp"
)

// Container represents a ScyllaDB container type used in the module
type Container struct {
	testcontainers.Container
}

// WithConfig sets the YAML config file as an [io.Reader] to be used for the ScyllaDB container
func WithConfig(r io.Reader) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			Reader:            r,
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

// WithAlternator enables the Alternator (DynamoDB Compatible API) service in the ScyllaDB container,
// using the default HTTP port 8000.
// It will set the "alternator-port" parameter to the specified port.
// It will also set the "alternator-write-isolation" parameter to "always" as a command line argument to the container.
func WithAlternator() testcontainers.CustomizeRequestOption {
	portFlagValue := strings.ReplaceAll(alternatorPort, "/tcp", "")

	return func(req *testcontainers.GenericContainerRequest) error {
		req.ExposedPorts = append(req.ExposedPorts, alternatorPort)
		req.WaitingFor = wait.ForAll(req.WaitingFor, wait.ForListeningPort(alternatorPort))
		setCommandFlag(req, map[string]string{
			"--alternator-port":            portFlagValue,
			"--alternator-write-isolation": "always",
		})

		return nil
	}
}

// WithCustomCommands sets custom commands with values for the ScyllaDB container.
// Pass the command and the values as a list of strings in the following format: "--flag1=value", "--flag2", etc.
// In case of an invalid flag (not starting with "--" or "-"), this option returns an error,
// not applying any changes to the command line. Else, flags that exist in the command line overwrite the default commands.
// See more in the [ScyllaDB docs].
//
// [ScyllaDB docs]: https://opensource.docs.scylladb.com/stable/operating-scylla/procedures/tips/best-practices-scylla-on-docker.html
func WithCustomCommands(flags ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		flagsMap := make(map[string]string, len(flags))
		for _, flag := range flags {
			if !strings.HasPrefix(flag, "--") && !strings.HasPrefix(flag, "-") {
				return fmt.Errorf("invalid flag: %s", flag)
			}

			before, after, found := strings.Cut(flag, "=")
			if found {
				flagsMap[before] = after
			} else {
				flagsMap[flag] = ""
			}
		}

		setCommandFlag(req, flagsMap)
		return nil
	}
}

// ShardAwareConnectionHost returns the host and port of the ScyllaDB container with the shard-aware port
func (c Container) ShardAwareConnectionHost(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, shardAwarePort, "")
}

// NonShardAwareConnectionHost returns the host and port of the ScyllaDB container with the non-shard-aware port
func (c Container) NonShardAwareConnectionHost(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, port, "")
}

// AlternatorConnectionHost returns the host and port of the ScyllaDB container with the alternator port
func (c Container) AlternatorConnectionHost(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, alternatorPort, "")
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

// setCommandFlag sets the flags in the command line.
// It takes the array of commands from the GenericContainerRequest and a map of flags,
// and checks if the flag is present in the command line, overriding the value if it is.
// If the flag is not present, it's added to the command line.
func setCommandFlag(req *testcontainers.GenericContainerRequest, flags map[string]string) {
	cmds := []string{}

	for _, cmd := range req.Cmd {
		before, _, hasEquals := strings.Cut(cmd, "=")
		val, ok := flags[before]
		if ok {
			if hasEquals {
				cmds = append(cmds, before+"="+val)
			} else {
				cmds = append(cmds, before)
			}
			// The flag is present in the command line, so it's removed from the flags map
			// to avoid adding it to the end of the command line.
			delete(flags, before)
		} else {
			cmds = append(cmds, cmd)
		}
	}

	// The extra flags not present in the command line are added to the end of the command line,
	// and this could be in any order.
	for key, val := range flags {
		if val == "" {
			cmds = append(cmds, key)
		} else {
			cmds = append(cmds, key+"="+val)
		}
	}

	req.Cmd = cmds
}
