package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	tcpDockerHost1234  = "tcp://127.0.0.1:1234"
	tcpDockerHost33293 = "tcp://127.0.0.1:33293"
	tcpDockerHost4711  = "tcp://127.0.0.1:4711"
)

// unset environment variables to avoid side effects
// execute this function before each test
func resetTestEnv(t *testing.T) {
	t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "")
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "")
	t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "")
}

func TestReadConfig(t *testing.T) {
	resetTestEnv(t)

	t.Run("Config is read just once", func(t *testing.T) {
		t.Cleanup(Reset)

		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support
		t.Setenv("DOCKER_HOST", "")
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

		config := Read()

		expected := Config{
			RyukDisabled: true,
			Host:         "", // docker socket is empty at the properties file
		}

		assert.Equal(t, expected, config)

		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "false")

		config = Read()
		assert.Equal(t, expected, config)
	})
}

func TestReadTCConfig(t *testing.T) {
	resetTestEnv(t)

	const defaultHubPrefix string = "registry.mycompany.com/mirror"

	t.Run("HOME is not set", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support

		config := read()

		expected := Config{}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME is not set - TESTCONTAINERS_ env is set", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", defaultHubPrefix)
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")

		config := read()

		expected := Config{
			HubImageNamePrefix: defaultHubPrefix,
			RyukDisabled:       true,
			RyukPrivileged:     true,
			Host:               "", // docker socket is empty at the properties file
		}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir) // Windows support

		config := read()

		expected := Config{}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file - DOCKER_HOST env is set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir) // Windows support
		t.Setenv("DOCKER_HOST", tcpDockerHost33293)

		config := read()
		expected := Config{} // the config does not read DOCKER_HOST, that's why it's empty

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file - TESTCONTAINERS_ env is set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir) // Windows support
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", defaultHubPrefix)
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")

		config := read()
		expected := Config{
			HubImageNamePrefix: defaultHubPrefix,
			RyukDisabled:       true,
			RyukPrivileged:     true,
		}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME contains TC properties file", func(t *testing.T) {
		defaultRyukConnectionTimeout := 60 * time.Second
		defaultRyukReonnectionTimeout := 10 * time.Second
		defaultConfig := Config{
			RyukConnectionTimeout:   defaultRyukConnectionTimeout,
			RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
		}

		tests := []struct {
			name     string
			content  string
			env      map[string]string
			expected Config
		}{
			{
				"Single Docker host with spaces",
				"docker.host = " + tcpDockerHost33293,
				map[string]string{},
				Config{
					Host:                    tcpDockerHost33293,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"Multiple docker host entries, last one wins",
				`docker.host = ` + tcpDockerHost33293 + `
	docker.host = ` + tcpDockerHost4711 + `
	`,
				map[string]string{},
				Config{
					Host:                    tcpDockerHost4711,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"Multiple docker host entries, last one wins, with TLS",
				`docker.host = ` + tcpDockerHost33293 + `
	docker.host = ` + tcpDockerHost4711 + `
	docker.host = ` + tcpDockerHost1234 + `
	docker.tls.verify = 1
	`,
				map[string]string{},
				Config{
					Host:                    tcpDockerHost1234,
					TLSVerify:               1,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"Empty file",
				"",
				map[string]string{},
				Config{
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"Non-valid properties are ignored",
				`foo = bar
	docker.host = ` + tcpDockerHost1234 + `
			`,
				map[string]string{},
				Config{
					Host:                    tcpDockerHost1234,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"Single Docker host without spaces",
				"docker.host=" + tcpDockerHost33293,
				map[string]string{},
				Config{
					Host:                    tcpDockerHost33293,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"Comments are ignored",
				`#docker.host=` + tcpDockerHost33293,
				map[string]string{},
				defaultConfig,
			},
			{
				"Multiple docker host entries, last one wins, with TLS and cert path",
				`#docker.host = ` + tcpDockerHost33293 + `
	docker.host = ` + tcpDockerHost4711 + `
	docker.host = ` + tcpDockerHost1234 + `
	docker.cert.path=/tmp/certs`,
				map[string]string{},
				Config{
					Host:                    tcpDockerHost1234,
					CertPath:                "/tmp/certs",
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Ryuk disabled using properties",
				`ryuk.disabled=true`,
				map[string]string{},
				Config{
					RyukDisabled:            true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Ryuk container privileged using properties",
				`ryuk.container.privileged=true`,
				map[string]string{},
				Config{
					RyukPrivileged:          true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Ryuk container timeouts configured using properties",
				`ryuk.connection.timeout=12s
	ryuk.reconnection.timeout=13s`,
				map[string]string{},
				Config{
					RyukReconnectionTimeout: 13 * time.Second,
					RyukConnectionTimeout:   12 * time.Second,
				},
			},
			{
				"With Ryuk disabled using an env var",
				``,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				Config{
					RyukDisabled:            true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Ryuk container privileged using an env var",
				``,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				Config{
					RyukPrivileged:          true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Ryuk disabled using an env var and properties. Env var wins (0)",
				`ryuk.disabled=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				Config{
					RyukDisabled:            true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Ryuk disabled using an env var and properties. Env var wins (1)",
				`ryuk.disabled=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				Config{
					RyukDisabled:            true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Ryuk disabled using an env var and properties. Env var wins (2)",
				`ryuk.disabled=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "false",
				},
				defaultConfig,
			},
			{
				"With Ryuk disabled using an env var and properties. Env var wins (3)",
				`ryuk.disabled=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "false",
				},
				defaultConfig,
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var wins (0)",
				`ryuk.container.privileged=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				Config{
					RyukPrivileged:          true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var wins (1)",
				`ryuk.container.privileged=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				Config{
					RyukPrivileged:          true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var wins (2)",
				`ryuk.container.privileged=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "false",
				},
				defaultConfig,
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var wins (3)",
				`ryuk.container.privileged=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "false",
				},
				defaultConfig,
			},
			{
				"With TLS verify using properties when value is wrong",
				`ryuk.container.privileged=false
				docker.tls.verify = ERROR`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED":             "true",
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				Config{
					RyukDisabled:   true,
					RyukPrivileged: true,
				},
			},
			{
				"With Ryuk disabled using an env var and properties. Env var does not win because it's not a boolean value",
				`ryuk.disabled=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "foo",
				},
				defaultConfig,
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var does not win because it's not a boolean value",
				`ryuk.container.privileged=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "foo",
				},
				defaultConfig,
			},
			{
				"With Hub image name prefix set as a property",
				`hub.image.name.prefix=` + defaultHubPrefix + `/props/`,
				map[string]string{},
				Config{
					HubImageNamePrefix:      defaultHubPrefix + "/props/",
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Hub image name prefix set as env var",
				``,
				map[string]string{
					"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX": defaultHubPrefix + "/env/",
				},
				Config{
					HubImageNamePrefix:      defaultHubPrefix + "/env/",
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
			{
				"With Hub image name prefix set as env var and properties: Env var wins",
				`hub.image.name.prefix=` + defaultHubPrefix + `/props/`,
				map[string]string{
					"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX": defaultHubPrefix + "/env/",
				},
				Config{
					HubImageNamePrefix:      defaultHubPrefix + "/env/",
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReonnectionTimeout,
				},
			},
		}
		for _, tt := range tests {
			t.Run(fmt.Sprintf(tt.name), func(t *testing.T) {
				tmpDir := t.TempDir()
				t.Setenv("HOME", tmpDir)
				t.Setenv("USERPROFILE", tmpDir) // Windows support
				for k, v := range tt.env {
					t.Setenv(k, v)
				}
				if err := os.WriteFile(filepath.Join(tmpDir, ".testcontainers.properties"), []byte(tt.content), 0o600); err != nil {
					t.Errorf("Failed to create the file: %v", err)
					return
				}

				//
				config := read()

				assert.Equal(t, tt.expected, config, "Configuration doesn't not match")
			})
		}
	})
}
