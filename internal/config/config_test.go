package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	tcpDockerHost1234  = "tcp://127.0.0.1:1234"
	tcpDockerHost33293 = "tcp://127.0.0.1:33293"
	tcpDockerHost4711  = "tcp://127.0.0.1:4711"
)

// setupTestDir creates a temporary directory and sets it as HOME and USERPROFILE
func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir) // Windows support
	return tmpDir
}

// setupTestProperties writes the given content to a .testcontainers.properties file in the given directory
func setupTestProperties(t *testing.T, dir, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, ".testcontainers.properties"), []byte(content), 0o600)
	require.NoErrorf(t, err, "Failed to create the properties file")
}

// setupTestEnv sets up the test environment with the given environment variables
func setupTestEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}
}

// unset environment variables to avoid side effects
// execute this function before each test
func resetTestEnv(t *testing.T) {
	t.Helper()

	setupTestEnv(t, map[string]string{
		"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX":     "",
		"TESTCONTAINERS_RYUK_DISABLED":             "",
		"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "",
		"RYUK_VERBOSE":                        "",
		"RYUK_RECONNECTION_TIMEOUT":           "",
		"RYUK_CONNECTION_TIMEOUT":             "",
		"TESTCONTAINERS_PORT_MAPPING_TIMEOUT": "",
	})
}

// defaultConfig returns a Config with default timeout values used in tests.
// The default values are:
//   - RyukConnectionTimeout: 60 seconds
//   - RyukReconnectionTimeout: 10 seconds
//   - TestcontainersPortMappingTimeout: 5 seconds
func defaultConfig() Config {
	return Config{
		RyukConnectionTimeout:            60 * time.Second,
		RyukReconnectionTimeout:          10 * time.Second,
		TestcontainersPortMappingTimeout: 5 * time.Second,
	}
}

// defaultTestConfig creates a new Config with default values and applies the given overrides.
// It is the main helper function for creating test configurations. Each override function
// modifies a specific aspect of the default configuration.
//
// Example usage:
//
//	config := defaultTestConfig(
//	    withHost("tcp://localhost:1234"),
//	    withRyukDisabled(true),
//	    withTimeouts(12*time.Second, 13*time.Second, 11*time.Second),
//	)
//
// For custom modifications not covered by the standard helpers, you can use an inline function:
//
//	config := defaultTestConfig(
//	    withHost("tcp://localhost:1234"),
//	    func(c *Config) { c.TLSVerify = 1 },
//	)
func defaultTestConfig(overrides ...func(*Config)) Config {
	config := defaultConfig()
	for _, override := range overrides {
		override(&config)
	}
	return config
}

// withHost returns a function that sets the Docker host URL in a Config.
// This is used to specify the Docker daemon connection URL.
//
// Example:
//
//	config := defaultTestConfig(withHost("tcp://localhost:1234"))
func withHost(host string) func(*Config) {
	return func(c *Config) {
		c.Host = host
	}
}

// withRyukDisabled returns a function that sets whether Ryuk (the resource reaper) is disabled.
// When disabled, Ryuk will not be started and containers will not be automatically removed.
//
// Example:
//
//	config := defaultTestConfig(withRyukDisabled(true))
func withRyukDisabled(disabled bool) func(*Config) {
	return func(c *Config) {
		c.RyukDisabled = disabled
	}
}

// withRyukPrivileged returns a function that sets whether the Ryuk container should run in privileged mode.
// Privileged mode is required for some Docker operations but may have security implications.
//
// Example:
//
//	config := defaultTestConfig(withRyukPrivileged(true))
func withRyukPrivileged(privileged bool) func(*Config) {
	return func(c *Config) {
		c.RyukPrivileged = privileged
	}
}

// withRyukVerbose returns a function that sets whether Ryuk should run in verbose mode.
// Verbose mode enables additional logging for debugging purposes.
//
// Example:
//
//	config := defaultTestConfig(withRyukVerbose(true))
func withRyukVerbose(verbose bool) func(*Config) {
	return func(c *Config) {
		c.RyukVerbose = verbose
	}
}

// withTimeouts returns a function that sets all timeout-related fields in a Config.
// The timeouts control various aspects of container and resource management:
//   - connection: Time to wait for initial connection to Ryuk
//   - reconnection: Time to wait between reconnection attempts to Ryuk
//   - portMapping: Time to wait for port mapping operations
//
// Example:
//
//	config := defaultTestConfig(withTimeouts(12*time.Second, 13*time.Second, 11*time.Second))
func withTimeouts(connection, reconnection, portMapping time.Duration) func(*Config) {
	return func(c *Config) {
		c.RyukConnectionTimeout = connection
		c.RyukReconnectionTimeout = reconnection
		c.TestcontainersPortMappingTimeout = portMapping
	}
}

// withHubImagePrefix returns a function that sets the prefix for Docker Hub image names.
// This is used to configure a custom registry or mirror for pulling Docker images.
//
// Example:
//
//	config := defaultTestConfig(withHubImagePrefix("registry.mycompany.com/mirror"))
func withHubImagePrefix(prefix string) func(*Config) {
	return func(c *Config) {
		c.HubImageNamePrefix = prefix
	}
}

func TestReadConfig(t *testing.T) {
	resetTestEnv(t)

	t.Run("config-read-once", func(t *testing.T) {
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

		require.Equal(t, expected, config)

		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "false")
		config = Read()
		require.Equal(t, expected, config)
	})
}

func TestReadTCConfig(t *testing.T) {
	resetTestEnv(t)

	const defaultHubPrefix string = "registry.mycompany.com/mirror"

	// Group 1: Basic environment setup tests
	t.Run("environment-setup", func(t *testing.T) {
		t.Run("HOME-not-set", func(t *testing.T) {
			t.Setenv("HOME", "")
			t.Setenv("USERPROFILE", "") // Windows support

			config := read()
			require.Equal(t, Config{}, config)
		})

		t.Run("HOME-does-not-contain-TC-props-file", func(t *testing.T) {
			setupTestDir(t)
			config := read()
			require.Equal(t, Config{}, config)
		})

		t.Run("HOME-does-not-contain-TC-props-file-DOCKER_HOST-env-is-set", func(t *testing.T) {
			setupTestDir(t)
			t.Setenv("DOCKER_HOST", tcpDockerHost33293)
			config := read()
			require.Equal(t, Config{}, config) // the config does not read DOCKER_HOST
		})
	})

	// Group 2: Environment variables tests
	t.Run("environment-variables", func(t *testing.T) {
		t.Run("HOME-is-not-set-TESTCONTAINERS_env-is-set", func(t *testing.T) {
			t.Setenv("HOME", "")
			t.Setenv("USERPROFILE", "") // Windows support
			env := map[string]string{
				"TESTCONTAINERS_RYUK_DISABLED":             "true",
				"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX":     defaultHubPrefix,
				"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				"RYUK_RECONNECTION_TIMEOUT":                "13s",
				"RYUK_CONNECTION_TIMEOUT":                  "12s",
				"TESTCONTAINERS_PORT_MAPPING_TIMEOUT":      "11s",
			}
			setupTestEnv(t, env)

			expected := defaultTestConfig(
				withHubImagePrefix(defaultHubPrefix),
				withRyukDisabled(true),
				withRyukPrivileged(true),
				withTimeouts(12*time.Second, 13*time.Second, 11*time.Second),
			)
			require.Equal(t, expected, read())
		})

		t.Run("HOME-does-not-contain-TC-props-file-TESTCONTAINERS_env-is-set", func(t *testing.T) {
			setupTestDir(t)
			env := map[string]string{
				"TESTCONTAINERS_RYUK_DISABLED":             "true",
				"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX":     defaultHubPrefix,
				"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				"RYUK_VERBOSE":                        "true",
				"RYUK_RECONNECTION_TIMEOUT":           "13s",
				"RYUK_CONNECTION_TIMEOUT":             "12s",
				"TESTCONTAINERS_PORT_MAPPING_TIMEOUT": "11s",
			}
			setupTestEnv(t, env)

			expected := defaultTestConfig(
				withHubImagePrefix(defaultHubPrefix),
				withRyukDisabled(true),
				withRyukPrivileged(true),
				withRyukVerbose(true),
				withTimeouts(12*time.Second, 13*time.Second, 11*time.Second),
			)
			require.Equal(t, expected, read())
		})
	})

	// Group 3: Properties file tests
	t.Run("properties-file", func(t *testing.T) {
		// Group 3.1: Docker host configuration
		t.Run("docker-host", func(t *testing.T) {
			tests := []struct {
				name     string
				content  string
				expected Config
			}{
				{
					"single-docker-host-with-spaces",
					"docker.host = " + tcpDockerHost33293,
					defaultTestConfig(withHost(tcpDockerHost33293)),
				},
				{
					"single-docker-host-without-spaces",
					"docker.host=" + tcpDockerHost33293,
					defaultTestConfig(withHost(tcpDockerHost33293)),
				},
				{
					"multiple-docker-host-entries-last-one-wins",
					`docker.host = ` + tcpDockerHost33293 + `
docker.host = ` + tcpDockerHost4711,
					defaultTestConfig(withHost(tcpDockerHost4711)),
				},
				{
					"multiple-docker-host-entries-with-TLS",
					`docker.host = ` + tcpDockerHost33293 + `
docker.host = ` + tcpDockerHost4711 + `
docker.host = ` + tcpDockerHost1234 + `
docker.tls.verify = 1`,
					defaultTestConfig(
						withHost(tcpDockerHost1234),
						func(c *Config) { c.TLSVerify = 1 },
					),
				},
				{
					"multiple-docker-host-entries-with-TLS-and-cert-path",
					`#docker.host = ` + tcpDockerHost33293 + `
docker.host = ` + tcpDockerHost4711 + `
docker.host = ` + tcpDockerHost1234 + `
docker.cert.path=/tmp/certs`,
					defaultTestConfig(
						withHost(tcpDockerHost1234),
						func(c *Config) { c.CertPath = "/tmp/certs" },
					),
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					tmpDir := setupTestDir(t)
					setupTestProperties(t, tmpDir, tt.content)
					require.Equal(t, tt.expected, read())
				})
			}
		})

		// Group 3.2: Ryuk configuration
		t.Run("ryuk", func(t *testing.T) {
			tests := []struct {
				name     string
				content  string
				expected Config
			}{
				{
					"ryuk-disabled",
					"ryuk.disabled=true",
					defaultTestConfig(withRyukDisabled(true)),
				},
				{
					"ryuk-privileged",
					"ryuk.container.privileged=true",
					defaultTestConfig(withRyukPrivileged(true)),
				},
				{
					"ryuk-verbose",
					"ryuk.verbose=true",
					defaultTestConfig(withRyukVerbose(true)),
				},
				{
					"ryuk-timeouts",
					`ryuk.connection.timeout=12s
ryuk.reconnection.timeout=13s
tc.port.mapping.timeout=11s`,
					defaultTestConfig(withTimeouts(12*time.Second, 13*time.Second, 11*time.Second)),
				},
				{
					"port-mapping-timeout",
					"tc.port.mapping.timeout=14s",
					defaultTestConfig(withTimeouts(60*time.Second, 10*time.Second, 14*time.Second)),
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					tmpDir := setupTestDir(t)
					setupTestProperties(t, tmpDir, tt.content)
					require.Equal(t, tt.expected, read())
				})
			}
		})

		// Group 3.3: Hub image configuration
		t.Run("hub-image", func(t *testing.T) {
			tests := []struct {
				name     string
				content  string
				expected Config
			}{
				{
					"hub-image-prefix",
					"hub.image.name.prefix=" + defaultHubPrefix + "/props/",
					defaultTestConfig(withHubImagePrefix(defaultHubPrefix + "/props/")),
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					tmpDir := setupTestDir(t)
					setupTestProperties(t, tmpDir, tt.content)
					require.Equal(t, tt.expected, read())
				})
			}
		})

		// Group 3.4: Edge cases
		t.Run("edge-cases", func(t *testing.T) {
			tests := []struct {
				name     string
				content  string
				expected Config
			}{
				{
					"empty-file",
					"",
					defaultConfig(),
				},
				{
					"comments-are-ignored",
					"#docker.host=" + tcpDockerHost33293,
					defaultConfig(),
				},
				{
					"non-valid-properties-are-ignored",
					`foo = bar
docker.host = ` + tcpDockerHost1234,
					defaultTestConfig(withHost(tcpDockerHost1234)),
				},
				{
					"invalid-TLS-verify-value",
					`ryuk.container.privileged=false
docker.tls.verify = ERROR`,
					Config{
						// read() doesn't set default values, so timeouts should be zero
					},
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					tmpDir := setupTestDir(t)
					setupTestProperties(t, tmpDir, tt.content)
					require.Equal(t, tt.expected, read())
				})
			}
		})
	})

	// Group 4: environment-variables-vs-properties-precedence
	t.Run("precedence", func(t *testing.T) {
		tests := []struct {
			name     string
			content  string
			env      map[string]string
			expected Config
		}{
			{
				"ryuk-disabled-env-var-wins",
				"ryuk.disabled=false",
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				defaultTestConfig(withRyukDisabled(true)),
			},
			{
				"ryuk-privileged-env-var-wins",
				"ryuk.container.privileged=false",
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				defaultTestConfig(withRyukPrivileged(true)),
			},
			{
				"ryuk-verbose-env-var-wins",
				"ryuk.verbose=false",
				map[string]string{
					"RYUK_VERBOSE": "true",
				},
				defaultTestConfig(withRyukVerbose(true)),
			},
			{
				"ryuk-timeouts-env-vars-win",
				`ryuk.connection.timeout=22s
ryuk.reconnection.timeout=23s
tc.port.mapping.timeout=21s`,
				map[string]string{
					"RYUK_RECONNECTION_TIMEOUT":           "13s",
					"RYUK_CONNECTION_TIMEOUT":             "12s",
					"TESTCONTAINERS_PORT_MAPPING_TIMEOUT": "11s",
				},
				defaultTestConfig(withTimeouts(12*time.Second, 13*time.Second, 11*time.Second)),
			},
			{
				"hub-image-prefix-env-var-wins",
				"hub.image.name.prefix=" + defaultHubPrefix + "/props/",
				map[string]string{
					"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX": defaultHubPrefix + "/env/",
				},
				defaultTestConfig(withHubImagePrefix(defaultHubPrefix + "/env/")),
			},
			{
				"invalid-boolean-env-var-is-ignored",
				"ryuk.disabled=false",
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "foo",
				},
				defaultConfig(),
			},
			{
				"invalid-privileged-env-var-is-ignored",
				"ryuk.container.privileged=false",
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "foo",
				},
				defaultConfig(),
			},
			{
				"port-mapping-timeout-env-var-wins",
				"tc.port.mapping.timeout=22s",
				map[string]string{
					"TESTCONTAINERS_PORT_MAPPING_TIMEOUT": "14s",
				},
				defaultTestConfig(withTimeouts(60*time.Second, 10*time.Second, 14*time.Second)),
			},
			{
				"invalid-port-mapping-timeout-env-var-is-ignored",
				"tc.port.mapping.timeout=22s",
				map[string]string{
					"TESTCONTAINERS_PORT_MAPPING_TIMEOUT": "invalid",
				},
				defaultTestConfig(withTimeouts(60*time.Second, 10*time.Second, 22*time.Second)),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmpDir := setupTestDir(t)
				setupTestEnv(t, tt.env)
				setupTestProperties(t, tmpDir, tt.content)
				require.Equal(t, tt.expected, read())
			})
		}
	})
}
