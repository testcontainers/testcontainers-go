package redis

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type LogLevel string

const (
	// redisPort is the port for the Redis connection
	redisPort = "6379/tcp"

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
	schema := "redis"
	if c.settings.tlsEnabled {
		schema = "rediss"
	}
	return c.PortEndpoint(ctx, redisPort, schema)
}

// TLSConfig returns the TLS configuration for the Redis container, nil if TLS is not enabled.
func (c *RedisContainer) TLSConfig() *tls.Config {
	return c.settings.tlsConfig
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Redis container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RedisContainer, error) {
	return Run(ctx, "redis:7", opts...)
}

// Run creates an instance of the Redis container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*RedisContainer, error) {
	// Process custom options to extract settings
	var settings options
	for _, opt := range opts {
		if opt, ok := opt.(Option); ok {
			if err := opt(&settings); err != nil {
				return nil, err
			}
		}
	}

	waitStrategies := []wait.Strategy{
		wait.ForLog("* Ready to accept connections"),
	}

	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(redisPort),
	}

	if settings.tlsEnabled {
		// wait for the TLS port to be available
		waitStrategies = append(waitStrategies, wait.ForListeningPort(redisPort).WithStartupTimeout(time.Second*10))

		// Generate TLS certificates in the fly and add them to the container before it starts.
		// Update the CMD to use the TLS certificates.
		caCert, clientCert, serverCert, err := createTLSCerts()
		if err != nil {
			return nil, fmt.Errorf("create tls certs: %w", err)
		}

		// Update the CMD to use the TLS certificates.
		cmds := []string{
			"--tls-port", strings.Replace(redisPort, "/tcp", "", 1),
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
				},
			),
		)

		settings.tlsConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			RootCAs:      caCert.TLSConfig().RootCAs,
			Certificates: clientCert.TLSConfig().Certificates,
			ServerName:   "localhost", // Match the server cert's common name
		}
	}

	moduleOpts = append(moduleOpts, testcontainers.WithWaitStrategy(waitStrategies...))

	// Append the customizers passed to the Run function.
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *RedisContainer
	if ctr != nil {
		c = &RedisContainer{Container: ctr, settings: settings}
	}

	if err != nil {
		return c, fmt.Errorf("run redis: %w", err)
	}

	return c, nil
}
