package testcontainers

import (
	"github.com/testcontainers/testcontainers-go/internal/config"
)

// Config represents the configuration for Testcontainers
type Config = config.Config

// ReadConfig reads from testcontainers properties file, storing the result in a singleton instance
// of the TestcontainersConfig struct
func ReadConfig() Config {
	return config.Read()
}
