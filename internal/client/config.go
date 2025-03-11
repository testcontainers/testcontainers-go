package client

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/magiconair/properties"
)

// config represents the configuration for Testcontainers.
// User values are read from ~/.testcontainers.properties file which can be overridden
// using the specified environment variables. For more information, see [Custom Configuration].
//
// The Ryuk fields controls the [Garbage Collector] feature, which ensures that resources are
// cleaned up after the test execution.
//
// [Garbage Collector]: https://golang.testcontainers.org/features/garbage_collector/
// [Custom Configuration]: https://golang.testcontainers.org/features/configuration/
type config struct { // TODO: consider renaming adding default values to the struct fields.
	// Host is the address of the Docker daemon.
	Host string `properties:"docker.host" env:"DOCKER_HOST"`

	// TLSVerify is a flag to enable or disable TLS verification when connecting to a Docker daemon.
	TLSVerify bool `properties:"docker.tls.verify" env:"DOCKER_TLS_VERIFY"`

	// CertPath is the path to the directory containing the Docker certificates.
	// This is used when connecting to a Docker daemon over TLS.
	CertPath string `properties:"docker.cert.path" env:"DOCKER_CERT_PATH"`

	// HubImageNamePrefix is the prefix used for the images pulled from the Docker Hub.
	// This is useful when running tests in environments with restricted internet access.
	HubImageNamePrefix string `properties:"hub.image.name.prefix" env:"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX"`

	// TestcontainersHost is the address of the Testcontainers host.
	TestcontainersHost string `properties:"tc.host" env:"TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE"`

	// Ryuk is the configuration for the Garbage Collector.
	Ryuk ryukConfig
}

type ryukConfig struct {
	// Disabled is a flag to enable or disable the Garbage Collector.
	// Setting this to true will prevent testcontainers from automatically cleaning up
	// resources, which is particularly important in tests which timeout as they
	// don't run test clean up.
	Disabled bool `properties:"ryuk.disabled" env:"TESTCONTAINERS_RYUK_DISABLED"`

	// Privileged is a flag to enable or disable the privileged mode for the Garbage Collector container.
	// Setting this to true will run the Garbage Collector container in privileged mode.
	Privileged bool `properties:"ryuk.container.privileged" env:"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED"`

	// ReconnectionTimeout is the time to wait before attempting to reconnect to the Garbage Collector container.
	ReconnectionTimeout time.Duration `properties:"ryuk.reconnection.timeout,default=10s" env:"TESTCONTAINERS_RYUK_RECONNECTION_TIMEOUT"`

	// ConnectionTimeout is the time to wait before timing out when connecting to the Garbage Collector container.
	ConnectionTimeout time.Duration `properties:"ryuk.connection.timeout,default=1m" env:"TESTCONTAINERS_RYUK_CONNECTION_TIMEOUT"`

	// Verbose is a flag to enable or disable verbose logging for the Garbage Collector.
	Verbose bool `properties:"ryuk.verbose" env:"TESTCONTAINERS_RYUK_VERBOSE"`
}

// newConfig returns a new configuration loaded from the properties file
// located in the user's home directory and overridden by environment variables.
func newConfig() (*config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home dir: %w", err)
	}

	props, err := properties.LoadFiles([]string{filepath.Join(home, ".testcontainers.properties")}, properties.UTF8, true)
	if err != nil {
		return nil, fmt.Errorf("load properties file: %w", err)
	}

	var cfg config
	if err := props.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode properties: %w", err)
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	return &cfg, nil
}
