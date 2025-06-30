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
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{redisPort},
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
		wait.ForListeningPort(redisPort).WithStartupTimeout(time.Second * 10),
		wait.ForLog("* Ready to accept connections"),
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
	var c *RedisContainer
	if container != nil {
		c = &RedisContainer{Container: container, settings: settings}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
