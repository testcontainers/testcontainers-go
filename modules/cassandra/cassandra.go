package cassandra

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	port = nat.Port("9042/tcp")
)

// CassandraContainer represents the Cassandra container type used in the module
type CassandraContainer struct {
	testcontainers.Container
}

// ConnectionHost returns the host and port of the cassandra container, using the default, native 9000 port, and
// obtaining the host and exposed port from the container
func (c *CassandraContainer) ConnectionHost(ctx context.Context) (string, error) {
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

// WithConfigFile sets the YAML config file to be used for the cassandra container
// It will also set the "configFile" parameter to the path of the config file
// as a command line argument to the container.
func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/cassandra/cassandra.yaml",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

// WithInitScripts sets the init cassandra queries to be run when the container starts
func WithInitScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		var initScripts []testcontainers.ContainerFile
		var execs []testcontainers.Executable
		for _, script := range scripts {
			cf := testcontainers.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: "/" + filepath.Base(script),
				FileMode:          0o755,
			}
			initScripts = append(initScripts, cf)

			execs = append(execs, initScript{File: cf.ContainerFilePath})
		}

		req.Files = append(req.Files, initScripts...)
		return testcontainers.WithAfterReadyCommand(execs...)(req)
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Cassandra container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CassandraContainer, error) {
	return Run(ctx, "cassandra:4.1.3", opts...)
}

// Run creates an instance of the Cassandra container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*CassandraContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{string(port)},
		Env: map[string]string{
			"CASSANDRA_SNITCH":          "GossipingPropertyFileSnitch",
			"JVM_OPTS":                  "-Dcassandra.skip_wait_for_gossip_to_settle=0 -Dcassandra.initial_token=0",
			"HEAP_NEWSIZE":              "128M",
			"MAX_HEAP_SIZE":             "1024M",
			"CASSANDRA_ENDPOINT_SNITCH": "GossipingPropertyFileSnitch",
			"CASSANDRA_DC":              "datacenter1",
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
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *CassandraContainer
	if container != nil {
		c = &CassandraContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
