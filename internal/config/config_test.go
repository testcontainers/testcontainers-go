package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core/bootstrap"
)

const (
	tcpDockerHost1234  = "tcp://127.0.0.1:1234"
	tcpDockerHost33293 = "tcp://127.0.0.1:33293"
	tcpDockerHost4711  = "tcp://127.0.0.1:4711"
)

// unset environment variables to avoid side effects
// execute this function before each test
func resetTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "")
	t.Setenv("TESTCONTAINERS_SESSION_ID", "")
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "")
	t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "")
	t.Setenv("RYUK_VERBOSE", "")
	t.Setenv("RYUK_RECONNECTION_TIMEOUT", "")
	t.Setenv("RYUK_CONNECTION_TIMEOUT", "")
	t.Setenv("TESTCONTAINERS_STARTUP_TIMEOUT", "")
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
			SessionID:    bootstrap.SessionID(),
			RyukDisabled: true,
			Host:         "", // docker socket is empty at the properties file
		}

		require.Equal(t, expected, config)

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

		expected := Config{
			SessionID: bootstrap.SessionID(),
		}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME is not set - TESTCONTAINERS_ env is set", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", defaultHubPrefix)
		t.Setenv("TESTCONTAINERS_SESSION_ID", "foo")
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")
		t.Setenv("RYUK_RECONNECTION_TIMEOUT", "13s")
		t.Setenv("RYUK_CONNECTION_TIMEOUT", "12s")
		t.Setenv("TESTCONTAINERS_STARTUP_TIMEOUT", "3m")

		config := read()

		expected := Config{
			HubImageNamePrefix:      defaultHubPrefix,
			SessionID:               "foo",
			RyukDisabled:            true,
			RyukPrivileged:          true,
			Host:                    "", // docker socket is empty at the properties file
			RyukReconnectionTimeout: 13 * time.Second,
			RyukConnectionTimeout:   12 * time.Second,
			StartupTimeout:          3 * time.Minute,
		}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir) // Windows support

		config := read()

		expected := Config{
			SessionID: bootstrap.SessionID(),
		}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file - DOCKER_HOST env is set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir) // Windows support
		t.Setenv("DOCKER_HOST", tcpDockerHost33293)

		config := read()
		expected := Config{
			Host:      "", // the config does not read env var `DOCKER_HOST`, that's why `Host` it's empty
			SessionID: bootstrap.SessionID(),
		}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file - TESTCONTAINERS_ env is set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir) // Windows support
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", defaultHubPrefix)
		t.Setenv("TESTCONTAINERS_SESSION_ID", "foo")
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")
		t.Setenv("RYUK_VERBOSE", "true")
		t.Setenv("RYUK_RECONNECTION_TIMEOUT", "13s")
		t.Setenv("RYUK_CONNECTION_TIMEOUT", "12s")

		config := read()
		expected := Config{
			HubImageNamePrefix:      defaultHubPrefix,
			SessionID:               "foo",
			RyukDisabled:            true,
			RyukPrivileged:          true,
			RyukVerbose:             true,
			RyukReconnectionTimeout: 13 * time.Second,
			RyukConnectionTimeout:   12 * time.Second,
		}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME contains TC properties file", func(t *testing.T) {
		defaultRyukConnectionTimeout := 60 * time.Second
		defaultRyukReconnectionTimeout := 10 * time.Second
		defaultStartupTimeout := time.Minute
		defaultConfig := Config{
			SessionID:               bootstrap.SessionID(),
			RyukConnectionTimeout:   defaultRyukConnectionTimeout,
			RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
			StartupTimeout:          defaultStartupTimeout,
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
					SessionID:               bootstrap.SessionID(),
					Host:                    tcpDockerHost33293,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"Multiple docker host entries, last one wins",
				`docker.host = ` + tcpDockerHost33293 + `
	docker.host = ` + tcpDockerHost4711 + `
	`,
				map[string]string{},
				Config{
					SessionID:               bootstrap.SessionID(),
					Host:                    tcpDockerHost4711,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
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
					SessionID:               bootstrap.SessionID(),
					Host:                    tcpDockerHost1234,
					TLSVerify:               1,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"Empty file",
				"",
				map[string]string{},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"Non-valid properties are ignored",
				`foo = bar
	docker.host = ` + tcpDockerHost1234 + `
			`,
				map[string]string{},
				Config{
					SessionID:               bootstrap.SessionID(),
					Host:                    tcpDockerHost1234,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"Single Docker host without spaces",
				"docker.host=" + tcpDockerHost33293,
				map[string]string{},
				Config{
					SessionID:               bootstrap.SessionID(),
					Host:                    tcpDockerHost33293,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
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
					SessionID:               bootstrap.SessionID(),
					Host:                    tcpDockerHost1234,
					CertPath:                "/tmp/certs",
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk disabled using properties",
				`ryuk.disabled=true`,
				map[string]string{},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukDisabled:            true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk container privileged using properties",
				`ryuk.container.privileged=true`,
				map[string]string{},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukPrivileged:          true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk container timeouts configured using properties",
				`ryuk.connection.timeout=12s
	ryuk.reconnection.timeout=13s`,
				map[string]string{},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukReconnectionTimeout: 13 * time.Second,
					RyukConnectionTimeout:   12 * time.Second,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk container timeouts configured using env vars",
				``,
				map[string]string{
					"RYUK_RECONNECTION_TIMEOUT": "13s",
					"RYUK_CONNECTION_TIMEOUT":   "12s",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukReconnectionTimeout: 13 * time.Second,
					RyukConnectionTimeout:   12 * time.Second,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With startup timeout configured using properties",
				`startup.timeout=2m`,
				map[string]string{},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          2 * time.Minute,
				},
			},
			{
				"With startup timeout configured using env var and properties. Env var wins",
				`startup.timeout=2m`,
				map[string]string{
					"TESTCONTAINERS_STARTUP_TIMEOUT": "3m",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          3 * time.Minute,
				},
			},
			{
				"With Ryuk container timeouts configured using env vars and properties. Env var wins",
				`ryuk.connection.timeout=22s
	ryuk.reconnection.timeout=23s`,
				map[string]string{
					"RYUK_RECONNECTION_TIMEOUT": "13s",
					"RYUK_CONNECTION_TIMEOUT":   "12s",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukReconnectionTimeout: 13 * time.Second,
					RyukConnectionTimeout:   12 * time.Second,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk verbose configured using properties",
				`ryuk.verbose=true`,
				map[string]string{},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukVerbose:             true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk disabled using an env var",
				``,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukDisabled:            true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk container privileged using an env var",
				``,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukPrivileged:          true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk disabled using an env var and properties. Env var wins (0)",
				`ryuk.disabled=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukDisabled:            true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk disabled using an env var and properties. Env var wins (1)",
				`ryuk.disabled=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukDisabled:            true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
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
				"With Ryuk verbose using an env var and properties. Env var wins (0)",
				`ryuk.verbose=true`,
				map[string]string{
					"RYUK_VERBOSE": "true",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukVerbose:             true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk verbose using an env var and properties. Env var wins (1)",
				`ryuk.verbose=false`,
				map[string]string{
					"RYUK_VERBOSE": "true",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukVerbose:             true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk verbose using an env var and properties. Env var wins (2)",
				`ryuk.verbose=true`,
				map[string]string{
					"RYUK_VERBOSE": "false",
				},
				defaultConfig,
			},
			{
				"With Ryuk verbose using an env var and properties. Env var wins (3)",
				`ryuk.verbose=false`,
				map[string]string{
					"RYUK_VERBOSE": "false",
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
					SessionID:               bootstrap.SessionID(),
					RyukPrivileged:          true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var wins (1)",
				`ryuk.container.privileged=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					RyukPrivileged:          true,
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
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
					SessionID:      bootstrap.SessionID(),
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
					SessionID:               bootstrap.SessionID(),
					HubImageNamePrefix:      defaultHubPrefix + "/props/",
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Hub image name prefix set as env var",
				``,
				map[string]string{
					"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX": defaultHubPrefix + "/env/",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					HubImageNamePrefix:      defaultHubPrefix + "/env/",
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Hub image name prefix set as env var and properties: Env var wins",
				`hub.image.name.prefix=` + defaultHubPrefix + `/props/`,
				map[string]string{
					"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX": defaultHubPrefix + "/env/",
				},
				Config{
					SessionID:               bootstrap.SessionID(),
					HubImageNamePrefix:      defaultHubPrefix + "/env/",
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			//
			{
				"With Session ID set as a property",
				`session.id=foo`,
				map[string]string{},
				Config{
					SessionID:               "foo",
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
			{
				"With Session ID set using an env var and properties. Env var wins",
				`session.id=bar`,
				map[string]string{
					"TESTCONTAINERS_SESSION_ID": "foo",
				},
				Config{
					SessionID:               "foo",
					RyukConnectionTimeout:   defaultRyukConnectionTimeout,
					RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
					StartupTimeout:          defaultStartupTimeout,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmpDir := t.TempDir()
				t.Setenv("HOME", tmpDir)
				t.Setenv("USERPROFILE", tmpDir) // Windows support
				for k, v := range tt.env {
					t.Setenv(k, v)
				}
				err := os.WriteFile(filepath.Join(tmpDir, ".testcontainers.properties"), []byte(tt.content), 0o600)
				require.NoErrorf(t, err, "Failed to create the file")

				//
				config := read()

				assert.Equal(t, tt.expected, config, "Configuration doesn't not match")
			})
		}
	})
}

func TestValidateSessionID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			name      string
			sessionID string
		}{
			{"generated", bootstrap.SessionID()},
			{"alphanumeric", "session42"},
			{"with-hyphens", "ci-pipeline-42"},
			{"with-underscores", "ci_pipeline_42"},
			{"with-dots", "ci.pipeline.42"},
			{"uuid", "9e0f4a1a-9c1e-4b7f-9d6a-2f1c3b4d5e6f"},
			{"single-char", "a"},
			{"max-length", strings.Repeat("a", maxContainerNameLen-len(reaperNamePrefix))},
			// the resulting container name is prefixed, so a session ID starting with
			// punctuation still produces a valid name, e.g. "reaper__session".
			{"leading-hyphen", "-session"},
			{"leading-dot", ".session"},
			{"leading-underscore", "_session"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				require.NoError(t, validateSessionID(tt.sessionID))
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		tests := []struct {
			name      string
			sessionID string
		}{
			{"empty", ""},
			{"slash", "team/ci"},
			{"space", "team ci"},
			{"colon", "team:ci"},
			{"too-long", strings.Repeat("a", maxContainerNameLen-len(reaperNamePrefix)+1)},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				require.Error(t, validateSessionID(tt.sessionID))
			})
		}
	})
}

func TestReadConfigSessionIDValidation(t *testing.T) {
	resetTestEnv(t)

	t.Run("env/invalid/panics", func(t *testing.T) {
		t.Cleanup(Reset)

		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support
		t.Setenv("TESTCONTAINERS_SESSION_ID", "team/ci")

		require.PanicsWithValue(t,
			`invalid TESTCONTAINERS_SESSION_ID value "team/ci": must contain only alphanumeric characters, dots, hyphens and underscores`,
			func() { read() },
		)
	})

	t.Run("env/valid/is-used", func(t *testing.T) {
		t.Cleanup(Reset)

		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support
		t.Setenv("TESTCONTAINERS_SESSION_ID", "ci-pipeline-42")

		require.Equal(t, "ci-pipeline-42", read().SessionID)
	})

	t.Run("properties/invalid/panics", func(t *testing.T) {
		t.Cleanup(Reset)

		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir) // Windows support

		err := os.WriteFile(filepath.Join(tmpDir, ".testcontainers.properties"), []byte("session.id=team/ci"), 0o600)
		require.NoError(t, err)

		require.PanicsWithValue(t,
			`invalid session.id property value "team/ci": must contain only alphanumeric characters, dots, hyphens and underscores`,
			func() { read() },
		)
	})

	t.Run("not-set/falls-back-to-generated", func(t *testing.T) {
		t.Cleanup(Reset)

		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support

		require.Equal(t, bootstrap.SessionID(), read().SessionID)
	})
}
