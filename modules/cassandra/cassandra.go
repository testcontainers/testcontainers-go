package cassandra

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
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
	useTLS bool
}

// ConnectionHost returns the host and port of the cassandra container, using the default, native port,
// obtaining the host and exposed port from the container
func (c *CassandraContainer) ConnectionHost(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	// Use the secure port if TLS is enabled
	portToUse := port
	if c.useTLS {
		portToUse = securePort
	}

	mappedPort, err := c.MappedPort(ctx, portToUse)
	if err != nil {
		return "", err
	}
	return host + ":" + mappedPort.Port(), nil
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

// SSLOptions contains the configuration options for setting up SSL/TLS
type SSLOptions struct {
	KeystorePath      string
	KeystorePassword  string
	CertPath          string
	RequireClientAuth bool
}

// WithSSL enables SSL/TLS support on the Cassandra container
func WithSSL(sslOpts SSLOptions) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		// Add the secure port to the exposed ports
		hasSecurePort := false
		for _, p := range req.ExposedPorts {
			if p == string(securePort) {
				hasSecurePort = true
				break
			}
		}
		if !hasSecurePort {
			req.ExposedPorts = append(req.ExposedPorts, string(securePort))
		}

		// Set SSL environment variables
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		// If keystore path is provided, copy it to the container
		if sslOpts.KeystorePath != "" {
			keystoreFile := testcontainers.ContainerFile{
				HostFilePath:      sslOpts.KeystorePath,
				ContainerFilePath: "/etc/cassandra/conf/keystore.jks",
				FileMode:          0o644,
			}
			req.Files = append(req.Files, keystoreFile)
		}

		if sslOpts.CertPath != "" {
			certFile := testcontainers.ContainerFile{
				HostFilePath:      sslOpts.CertPath,
				ContainerFilePath: "/etc/cassandra/conf/cassandra.crt",
				FileMode:          0o644,
			}
			req.Files = append(req.Files, certFile)
		}

		// Mark that SSL is enabled for later use
		req.Labels = mergeMap(req.Labels, map[string]string{"testcontainers.cassandra.ssl": "true"})

		return nil
	}
}

// WithTLS is an alias for WithSSL for user convenience
func WithTLS(opts SSLOptions) testcontainers.CustomizeRequestOption {
	return WithSSL(opts)
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
			"CASSANDRA_SNITCH":                 "GossipingPropertyFileSnitch",
			"JVM_OPTS":                         "-Dcassandra.skip_wait_for_gossip_to_settle=0 -Dcassandra.initial_token=0",
			"HEAP_NEWSIZE":                     "128M",
			"MAX_HEAP_SIZE":                    "1024M",
			"CASSANDRA_ENDPOINT_SNITCH":        "GossipingPropertyFileSnitch",
			"CASSANDRA_DC":                     "datacenter1",
			"CASSANDRA_SKIP_WAIT_FOR_GOSSIP":   "1",
			"CASSANDRA_START_NATIVE_TRANSPORT": "true",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(port),
			wait.ForExec([]string{"cqlsh", "-e", "SELECT bootstrapped FROM system.local"}).WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return strings.Contains(string(data), "COMPLETED")
			}).WithStartupTimeout(2*time.Minute),
		),
	}

	c := &CassandraContainer{}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	sslEnabled := false
	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
		// Detect if WithSSL/WithTLS was used by checking the label
		if genericContainerReq.Labels != nil && genericContainerReq.Labels["testcontainers.cassandra.ssl"] == "true" {
			sslEnabled = true
		}
	}

	// If SSL is enabled, add a TLS wait strategy for the keystore file and SSL CQL port
	if sslEnabled {
		genericContainerReq.WaitingFor = wait.ForAll(
			genericContainerReq.WaitingFor,
			wait.ForListeningPort(securePort).WithStartupTimeout(3*time.Minute),
		)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if container != nil {
		c.Container = container
		// Set useTLS based on sslEnabled
		c.useTLS = sslEnabled
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// mergeMap is a helper to merge two string maps
func mergeMap(a, b map[string]string) map[string]string {
	if a == nil && b == nil {
		return nil
	}
	result := make(map[string]string)
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}
