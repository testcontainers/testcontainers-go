package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/magiconair/properties"
)

const ReaperDefaultImage = "docker.io/testcontainers/ryuk:0.5.1"

var (
	tcConfig     Config
	tcConfigOnce *sync.Once = new(sync.Once)
)

// Config represents the configuration for Testcontainers
// testcontainersConfig {
type Config struct {
	Host                    string        `properties:"docker.host,default="`
	TLSVerify               int           `properties:"docker.tls.verify,default=0"`
	CertPath                string        `properties:"docker.cert.path,default="`
	HubImageNamePrefix      string        `properties:"hub.image.name.prefix,default="`
	RyukDisabled            bool          `properties:"ryuk.disabled,default=false"`
	RyukPrivileged          bool          `properties:"ryuk.container.privileged,default=false"`
	RyukReconnectionTimeout time.Duration `properties:"ryuk.reconnection.timeout,default=10s"`
	RyukConnectionTimeout   time.Duration `properties:"ryuk.connection.timeout,default=1m"`
	TestcontainersHost      string        `properties:"tc.host,default="`
}

// }

// Read reads from testcontainers properties file, if it exists
// it is possible that certain values get overridden when set as environment variables
func Read() Config {
	tcConfigOnce.Do(func() {
		tcConfig = read()

		if tcConfig.RyukDisabled {
			ryukDisabledMessage := `
**********************************************************************************************
Ryuk has been disabled for the current execution. This can cause unexpected behavior in your environment.
More on this: https://golang.testcontainers.org/features/garbage_collector/
**********************************************************************************************`
			fmt.Println(ryukDisabledMessage)
		}
	})

	return tcConfig
}

// Reset resets the singleton instance of the Config struct,
// allowing to read the configuration again.
// Handy for testing, so do not use it in production code
// This function is not thread-safe
func Reset() {
	tcConfigOnce = new(sync.Once)
}

func read() Config {
	config := Config{}

	applyEnvironmentConfiguration := func(config Config) Config {
		ryukDisabledEnv := os.Getenv("TESTCONTAINERS_RYUK_DISABLED")
		if parseBool(ryukDisabledEnv) {
			config.RyukDisabled = ryukDisabledEnv == "true"
		}

		hubImageNamePrefix := os.Getenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX")
		if hubImageNamePrefix != "" {
			config.HubImageNamePrefix = hubImageNamePrefix
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

	return applyEnvironmentConfiguration(config)
}

func parseBool(input string) bool {
	_, err := strconv.ParseBool(input)
	return err == nil
}
