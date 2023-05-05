package testcontainers

import (
	"context"
	"sync"

	"github.com/testcontainers/testcontainers-go/internal/config"
)

var tcConfig TestcontainersConfig
var tcConfigOnce *sync.Once = new(sync.Once)

// TestcontainersConfig represents the configuration for Testcontainers
type TestcontainersConfig struct {
	Host           string `properties:"docker.host,default="`                    // Deprecated: use Config.Host instead
	TLSVerify      int    `properties:"docker.tls.verify,default=0"`             // Deprecated: use Config.TLSVerify instead
	CertPath       string `properties:"docker.cert.path,default="`               // Deprecated: use Config.CertPath instead
	RyukDisabled   bool   `properties:"ryuk.disabled,default=false"`             // Deprecated: use Config.RyukDisabled instead
	RyukPrivileged bool   `properties:"ryuk.container.privileged,default=false"` // Deprecated: use Config.RyukPrivileged instead
	Config         config.Config
}

// ReadConfig reads from testcontainers properties file, storing the result in a singleton instance
// of the TestcontainersConfig struct
// Deprecated use ReadConfigWithContext instead
func ReadConfig() TestcontainersConfig {
	return ReadConfigWithContext(context.Background())
}

// ReadConfigWithContext reads from testcontainers properties file, storing the result in a singleton instance
// of the TestcontainersConfig struct
func ReadConfigWithContext(ctx context.Context) TestcontainersConfig {
	tcConfigOnce.Do(func() {
		cfg := config.Read(ctx)

		tcConfig.Config = cfg

		if cfg.RyukDisabled {
			ryukDisabledMessage := `
**********************************************************************************************
Ryuk has been disabled for the current execution. This can cause unexpected behavior in your environment.
More on this: https://golang.testcontainers.org/features/garbage_collector/
**********************************************************************************************`
			Logger.Printf(ryukDisabledMessage)
			Logger.Printf("\n%+v", cfg)
		}
	})

	return tcConfig
}
