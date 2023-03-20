package testcontainers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/magiconair/properties"
)

var tcConfig TestcontainersConfig
var tcConfigOnce *sync.Once = new(sync.Once)

// TestcontainersConfig represents the configuration for Testcontainers
// testcontainersConfig {
type TestcontainersConfig struct {
	Host           string `properties:"docker.host,default="`
	TLSVerify      int    `properties:"docker.tls.verify,default=0"`
	CertPath       string `properties:"docker.cert.path,default="`
	RyukDisabled   bool   `properties:"ryuk.disabled,default=false"`
	RyukPrivileged bool   `properties:"ryuk.container.privileged,default=false"`
}

// }

// ReadConfig reads from testcontainers properties file, storing the result in a singleton instance
// of the TestcontainersConfig struct
func ReadConfig() TestcontainersConfig {
	tcConfigOnce.Do(func() {
		tcConfig = readConfig()

		if tcConfig.RyukDisabled {
			ryukDisabledMessage := `
**********************************************************************************************
Ryuk has been disabled for the current execution. This can cause unexpected behavior in your environment.
More on this: https://golang.testcontainers.org/features/garbage_collector/
**********************************************************************************************`
			Logger.Printf(ryukDisabledMessage)
			Logger.Printf("\n%+v", tcConfig)
		}
	})

	return tcConfig
}

// readConfig reads from testcontainers properties file, if it exists
// it is possible that certain values get overridden when set as environment variables
func readConfig() TestcontainersConfig {
	config := TestcontainersConfig{}

	applyEnvironmentConfiguration := func(config TestcontainersConfig) TestcontainersConfig {
		if dockerHostEnv := os.Getenv("DOCKER_HOST"); dockerHostEnv != "" {
			config.Host = dockerHostEnv
		}
		if config.Host == "" {
			config.Host = "unix:///var/run/docker.sock"
		}

		ryukDisabledEnv := os.Getenv("TESTCONTAINERS_RYUK_DISABLED")
		if parseBool(ryukDisabledEnv) {
			config.RyukDisabled = ryukDisabledEnv == "true"
		}

		ryukPrivilegedEnv := os.Getenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED")
		if parseBool(ryukPrivilegedEnv) {
			config.RyukPrivileged = ryukPrivilegedEnv == "true"
		}

		return config
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return applyEnvironmentConfiguration(config)
	}

	tcProp := filepath.Join(home, ".testcontainers.properties")
	// init from a file
	properties, err := properties.LoadFile(tcProp, properties.UTF8)
	if err != nil {
		return applyEnvironmentConfiguration(config)
	}

	if err := properties.Decode(&config); err != nil {
		fmt.Printf("invalid testcontainers properties file, returning an empty Testcontainers configuration: %v\n", err)
		return applyEnvironmentConfiguration(config)
	}

	fmt.Printf("Testcontainers properties file has been found: %s\n", tcProp)

	return applyEnvironmentConfiguration(config)
}

func parseBool(input string) bool {
	if _, err := strconv.ParseBool(input); err == nil {
		return true
	}
	return false
}
