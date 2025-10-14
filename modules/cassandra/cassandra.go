package cassandra

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	port       = nat.Port("9042/tcp")
	securePort = nat.Port("9142/tcp") // Common port for SSL/TLS connections
)

// CassandraContainer represents the Cassandra container type used in the module
type CassandraContainer struct {
	testcontainers.Container
	settings Options
}

// ConnectionHost returns the host and port of the cassandra container, using the default, native port,
// obtaining the host and exposed port from the container
func (c *CassandraContainer) ConnectionHost(ctx context.Context) (string, error) {
	// Use the secure port if TLS is enabled
	portToUse := port
	if c.settings.tlsConfig != nil {
		portToUse = securePort
	}

	return c.PortEndpoint(ctx, portToUse, "")
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
		Image: img,
		Env: map[string]string{
			"CASSANDRA_SNITCH":          "GossipingPropertyFileSnitch",
			"JVM_OPTS":                  "-Dcassandra.skip_wait_for_gossip_to_settle=0 -Dcassandra.initial_token=0",
			"HEAP_NEWSIZE":              "128M",
			"MAX_HEAP_SIZE":             "1024M",
			"CASSANDRA_ENDPOINT_SNITCH": "GossipingPropertyFileSnitch",
			"CASSANDRA_DC":              "datacenter1",
		},
		ExposedPorts: []string{string(port)},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	var settings Options
	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	// Set up wait strategies
	waitStrategies := []wait.Strategy{
		wait.ForListeningPort(port),
		wait.ForExec([]string{"cqlsh", "-e", "SELECT bootstrapped FROM system.local"}).WithResponseMatcher(func(body io.Reader) bool {
			data, _ := io.ReadAll(body)
			return strings.Contains(string(data), "COMPLETED")
		}).WithStartupTimeout(1 * time.Minute),
	}

	// Add TLS wait strategy if TLS config exists
	if settings.tlsConfig != nil {
		waitStrategies = append(waitStrategies, wait.ForListeningPort(securePort).WithStartupTimeout(1*time.Minute))
	}

	// Apply wait strategy using the correct method
	if err := testcontainers.WithWaitStrategy(wait.ForAll(waitStrategies...)).Customize(&genericContainerReq); err != nil {
		return nil, err
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *CassandraContainer
	if container != nil {
		c = &CassandraContainer{Container: container, settings: settings}
	}

	if err != nil {
		return c, fmt.Errorf("run cassandra: %w", err)
	}

	return c, nil
}

// TLSConfig returns the TLS configuration for the Cassandra container, nil if TLS is not enabled.
func (c *CassandraContainer) TLSConfig() *tls.Config {
	if c.settings.tlsConfig == nil {
		return nil
	}
	return c.settings.tlsConfig.Config
}
