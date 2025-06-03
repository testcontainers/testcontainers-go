package valkey

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ValkeyContainer represents the Valkey container type used in the module
type ValkeyContainer struct {
	testcontainers.Container
	settings options
}

// valkeyServerProcess is the name of the valkey server process
const valkeyServerProcess = "valkey-server"

type LogLevel string

const (
	// valkeyPort is the port for the Valkey connection
	valkeyPort = "6379/tcp"

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
	mappedPort, err := c.MappedPort(ctx, valkeyPort)
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	schema := "redis"
	if c.settings.tlsEnabled {
		schema = "rediss"
	}

	uri := fmt.Sprintf("%s://%s:%s", schema, hostIP, mappedPort.Port())
	return uri, nil
}

// TLSConfig returns the TLS configuration for the Valkey container, nil if TLS is not enabled.
func (c *ValkeyContainer) TLSConfig() *tls.Config {
	return c.settings.tlsConfig
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
		ExposedPorts: []string{valkeyPort},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	var settings options
	for _, opt := range opts {
		if opt, ok := opt.(Option); ok {
			if err := opt(&settings); err != nil {
				return nil, err
			}
		}
	}

	tcOpts := []testcontainers.ContainerCustomizer{}

	waitStrategies := []wait.Strategy{
		wait.ForListeningPort(valkeyPort).WithStartupTimeout(time.Second * 10),
		wait.ForLog("* Ready to accept connections"),
	}

	if settings.tlsEnabled {
		// wait for the TLS port to be available
		waitStrategies = append(waitStrategies, wait.ForListeningPort(valkeyPort).WithStartupTimeout(time.Second*10))

		// Generate TLS certificates in the fly and add them to the container before it starts.
		// Update the CMD to use the TLS certificates.
		caCert, clientCert, serverCert, err := createTLSCerts()
		if err != nil {
			return nil, fmt.Errorf("create tls certs: %w", err)
		}

		// Update the CMD to use the TLS certificates.
		cmds := []string{
			"--tls-port", strings.Replace(valkeyPort, "/tcp", "", 1),
			// Disable the default port, as described in https://redis.io/docs/latest/operate/oss_and_stack/management/security/encryption/#running-manually
			"--port", "0",
			"--tls-cert-file", "/tls/server.crt",
			"--tls-key-file", "/tls/server.key",
			"--tls-ca-cert-file", "/tls/ca.crt",
			"--tls-auth-clients", "yes",
		}

		tcOpts = append(tcOpts, testcontainers.WithCmdArgs(cmds...)) // Append the default CMD with the TLS certificates.
		tcOpts = append(tcOpts, testcontainers.WithFiles(
			testcontainers.ContainerFile{
				Reader:            bytes.NewReader(caCert.Bytes),
				ContainerFilePath: "/tls/ca.crt",
				FileMode:          0o644,
			},
			testcontainers.ContainerFile{
				Reader:            bytes.NewReader(serverCert.Bytes),
				ContainerFilePath: "/tls/server.crt",
				FileMode:          0o644,
			},
			testcontainers.ContainerFile{
				Reader:            bytes.NewReader(serverCert.KeyBytes),
				ContainerFilePath: "/tls/server.key",
				FileMode:          0o644,
			}))

		settings.tlsConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			RootCAs:      caCert.TLSConfig().RootCAs,
			Certificates: clientCert.TLSConfig().Certificates,
			ServerName:   "localhost", // Match the server cert's common name
		}
	}

	tcOpts = append(tcOpts, testcontainers.WithWaitStrategy(waitStrategies...))

	// Append the customizers passed to the Run function.
	tcOpts = append(tcOpts, opts...)

	// Apply the testcontainers customizers.
	for _, opt := range tcOpts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *ValkeyContainer
	if container != nil {
		c = &ValkeyContainer{Container: container, settings: settings}
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
