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
	schema := "redis"
	if c.settings.tlsEnabled {
		schema = "rediss"
	}

	return c.PortEndpoint(ctx, valkeyPort, schema)
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
	// Process custom options first
	var settings options
	for _, opt := range opts {
		if opt, ok := opt.(Option); ok {
			if err := opt(&settings); err != nil {
				return nil, fmt.Errorf("apply option: %w", err)
			}
		}
	}

	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(valkeyPort),
	}

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

		moduleOpts = append(moduleOpts,
			testcontainers.WithCmdArgs(cmds...),
			testcontainers.WithFiles(
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
				}),
		)

		settings.tlsConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			RootCAs:      caCert.TLSConfig().RootCAs,
			Certificates: clientCert.TLSConfig().Certificates,
			ServerName:   "localhost", // Match the server cert's common name
		}
	}

	moduleOpts = append(moduleOpts, testcontainers.WithWaitStrategy(waitStrategies...))

	ctr, err := testcontainers.Run(ctx, img, append(moduleOpts, opts...)...)
	var c *ValkeyContainer
	if ctr != nil {
		c = &ValkeyContainer{Container: ctr, settings: settings}
	}

	if err != nil {
		return c, fmt.Errorf("run valkey: %w", err)
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

		if err := testcontainers.WithFiles(cf)(req); err != nil {
			return err
		}

		if len(req.Cmd) == 0 {
			return testcontainers.WithCmd(valkeyServerProcess, defaultConfigFile)(req)
		}

		// prepend the command to run the valkey server with the config file, which must be the first argument of the valkey server process
		if req.Cmd[0] == valkeyServerProcess {
			// just insert the config file, then the rest of the args
			return testcontainers.WithCmd(append([]string{valkeyServerProcess, defaultConfigFile}, req.Cmd[1:]...)...)(req)
		}

		// prepend the valkey server and the config file, then the rest of the args
		return testcontainers.WithCmd(append([]string{valkeyServerProcess, defaultConfigFile}, req.Cmd...)...)(req)
	}
}

// WithLogLevel sets the log level for the valkey server process
// See https://redis.io/docs/reference/modules/modules-api-ref/#redismodule_log for more information.
func WithLogLevel(level LogLevel) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		return processValkeyServerArgs(req, []string{"--loglevel", string(level)})
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
		return processValkeyServerArgs(req, []string{"--save", strconv.Itoa(seconds), strconv.Itoa(changedKeys)})
	}
}

func processValkeyServerArgs(req *testcontainers.GenericContainerRequest, args []string) error {
	if len(req.Cmd) == 0 {
		return testcontainers.WithCmd(append([]string{valkeyServerProcess}, args...)...)(req)
	}

	// If valkey server is already set as the first argument, just append the args
	if req.Cmd[0] == valkeyServerProcess {
		return testcontainers.WithCmdArgs(args...)(req)
	}

	// valkey server is not set as the first argument, so prepend it alongside the existing command and args
	return testcontainers.WithCmd(append(append([]string{valkeyServerProcess}, req.Cmd...), args...)...)(req)
}
