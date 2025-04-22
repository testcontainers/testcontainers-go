package redis

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

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
	settings options
}

// ConnectionString returns the connection string for the Redis container.
// It uses the default 6379 port.
func (c *RedisContainer) ConnectionString(ctx context.Context) (string, error) {
	return c.connectionString(ctx, "6379/tcp")
}

// ConnectionStringTLS returns the connection string for the Redis container using TLS.
// It uses the TLS port defined in the options.
func (c *RedisContainer) ConnectionStringTLS(ctx context.Context) (string, error) {
	return c.connectionString(ctx, nat.Port(c.settings.tlsPort))
}

// TLSConfig returns the TLS configuration for the Redis container, nil if TLS is not enabled.
func (c *RedisContainer) TLSConfig() *tls.Config {
	return c.settings.tlsConfig
}

func (c *RedisContainer) connectionString(ctx context.Context, port nat.Port) (string, error) {
	mappedPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	schema := "redis://"
	if c.settings.withSecureURL {
		schema = "rediss://"
	}

	uri := fmt.Sprintf("%s%s:%s", schema, hostIP, mappedPort.Port())
	return uri, nil
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Redis container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RedisContainer, error) {
	return Run(ctx, "redis:7", opts...)
}

// Run creates an instance of the Redis container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*RedisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"6379/tcp"},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	settings := options{}
	for _, opt := range opts {
		if opt, ok := opt.(Option); ok {
			if err := opt(&settings); err != nil {
				return nil, err
			}
		}
	}

	tcOpts := []testcontainers.ContainerCustomizer{}

	waitStrategies := []wait.Strategy{
		wait.ForListeningPort("6379/tcp").WithStartupTimeout(time.Second * 10),
		wait.ForLog("* Ready to accept connections"),
	}

	if settings.tlsPort != "" {
		// wait for the TLS port to be available
		waitStrategies = append(waitStrategies, wait.ForListeningPort(nat.Port(settings.tlsPort)).WithStartupTimeout(time.Second*10))

		// Create a temporary directory to store the TLS certificates.
		tmpDir := os.TempDir()

		// Generate TLS certificates in the fly and add them to the container before it starts.
		// Update the CMD to use the TLS certificates.
		caCert, clientCert, serverCert := createTLSCerts(tmpDir)

		// Update the CMD to use the TLS certificates.
		cmds := []string{
			"--port", "6379",
			"--tls-port", strings.Replace(settings.tlsPort, "/tcp", "", 1),
			"--tls-cert-file", "/tls/server.crt",
			"--tls-key-file", "/tls/server.key",
			"--tls-ca-cert-file", "/tls/ca.crt",
			"--tls-auth-clients", "yes",
		}

		if settings.withMTLSDisabled {
			cmds = append(cmds, "--tls-auth-clients", "no")
		}

		tcOpts = append(tcOpts, testcontainers.WithCmd(cmds...)) // Replace the default CMD with the TLS certificates.
		tcOpts = append(tcOpts, testcontainers.WithExposedPorts(settings.tlsPort))
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

	// Apply the testcontainers customizers.
	for _, opt := range tcOpts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *RedisContainer
	if container != nil {
		c = &RedisContainer{Container: container, settings: settings}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

func processRedisServerArgs(req *testcontainers.GenericContainerRequest, args []string) {
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
