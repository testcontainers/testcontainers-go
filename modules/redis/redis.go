package redis

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type LogLevel string

const (
	// tlsPort is the port for the TLS connection
	tlsPort = "6380/tcp"

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
	return c.connectionString(ctx, nat.Port(redisPort))
}

// ConnectionStringTLS returns the connection string for the Redis container using TLS.
// It returns an error if TLS is not enabled, else the TLS port defined in the options
// is used to build the connection string.
func (c *RedisContainer) ConnectionStringTLS(ctx context.Context) (string, error) {
	if !c.settings.tlsEnabled {
		return "", errors.New("TLS is not enabled")
	}

	return c.connectionString(ctx, nat.Port(tlsPort))
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

	schema := "redis"
	if c.settings.tlsEnabled {
		schema = "rediss"
	}

	uri := fmt.Sprintf("%s://%s:%s", schema, hostIP, mappedPort.Port())
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
		ExposedPorts: []string{redisPort},
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
		wait.ForListeningPort(redisPort).WithStartupTimeout(time.Second * 10),
		wait.ForLog("* Ready to accept connections"),
	}

	if settings.tlsEnabled {
		// wait for the TLS port to be available
		waitStrategies = append(waitStrategies, wait.ForListeningPort(nat.Port(tlsPort)).WithStartupTimeout(time.Second*10))

		// Create a temporary directory to store the TLS certificates.
		tmpDir := os.TempDir()

		// Generate TLS certificates in the fly and add them to the container before it starts.
		// Update the CMD to use the TLS certificates.
		caCert, clientCert, serverCert := createTLSCerts(tmpDir)

		// Update the CMD to use the TLS certificates.
		cmds := []string{
			"--tls-port", strings.Replace(tlsPort, "/tcp", "", 1),
			"--tls-cert-file", "/tls/server.crt",
			"--tls-key-file", "/tls/server.key",
			"--tls-ca-cert-file", "/tls/ca.crt",
			"--tls-auth-clients", "yes",
		}

		tcOpts = append(tcOpts, testcontainers.WithCmdArgs(cmds...)) // Append the default CMD with the TLS certificates.
		tcOpts = append(tcOpts, testcontainers.WithExposedPorts(tlsPort))
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
