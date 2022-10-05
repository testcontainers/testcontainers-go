package properties

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/magiconair/properties"
)

var tcConfig *TestcontainersConfig
var tcConfigOnce sync.Once

// or through Decode
type TestcontainersConfig struct {
	Host           string `properties:"docker.host,default="`
	TLSVerify      int    `properties:"docker.tls.verify,default=0"`
	CertPath       string `properties:"docker.cert.path,default="`
	RyukPrivileged bool   `properties:"ryuk.container.privileged,default=false"`
}

// Get initializes the configuration in a lazy manner
func Get() *TestcontainersConfig {
	if tcConfig != nil {
		tcConfigOnce.Do(func() {
			tcConfig = configureTC()
		})
	}

	return tcConfig
}

// configureTC reads from testcontainers properties file, if it exists
// it is possible that certain values get overridden when set as environment variables
func configureTC() *TestcontainersConfig {
	applyEnvironmentConfiguration := func(config *TestcontainersConfig) *TestcontainersConfig {
		ryukPrivilegedEnv := os.Getenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED")
		if ryukPrivilegedEnv != "" {
			config.RyukPrivileged = ryukPrivilegedEnv == "true"
		}

		return config
	}

	config := &TestcontainersConfig{}

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

	if err := properties.Decode(config); err != nil {
		fmt.Printf("invalid testcontainers properties file, returning an empty Testcontainers configuration: %v\n", err)
	}

	return applyEnvironmentConfiguration(config)
}
