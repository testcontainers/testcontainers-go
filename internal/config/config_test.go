package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"dario.cat/mergo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "")
	t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "")
	t.Setenv("RYUK_VERBOSE", "")
	t.Setenv("RYUK_RECONNECTION_TIMEOUT", "")
	t.Setenv("RYUK_CONNECTION_TIMEOUT", "")
}

func TestReadConfig(t *testing.T) {
	resetTestEnv(t)

	t.Run("config/read-once", func(t *testing.T) {
		t.Cleanup(Reset)

		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support
		t.Setenv("DOCKER_HOST", "")
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

		config, err := Read()
		require.NoError(t, err)

		expected := Config{
			RyukDisabled: true,
			Host:         "", // docker socket is empty at the properties file
		}

		require.Equal(t, expected, config)

		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "false")

		config, err = Read()
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})
}

func TestReadTCConfig(t *testing.T) {
	resetTestEnv(t)

	const defaultHubPrefix string = "registry.mycompany.com/mirror"

	t.Run("home/unset", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support

		config, err := read()
		require.NoError(t, err)
		expected := Config{}

		assert.Equal(t, expected, config)
	})

	t.Run("home/unset/testcontainers-env/set", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "") // Windows support
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", defaultHubPrefix)
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")
		t.Setenv("RYUK_RECONNECTION_TIMEOUT", "13s")
		t.Setenv("RYUK_CONNECTION_TIMEOUT", "12s")

		config, err := read()
		require.NoError(t, err)

		expected := Config{
			HubImageNamePrefix:      defaultHubPrefix,
			RyukDisabled:            true,
			RyukPrivileged:          true,
			Host:                    "", // docker socket is empty at the properties file
			RyukReconnectionTimeout: 13 * time.Second,
			RyukConnectionTimeout:   12 * time.Second,
		}

		assert.Equal(t, expected, config)
	})

	t.Run("home/set/no-testcontainers-props-file", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir) // Windows support

		config, err := read()
		require.NoError(t, err)

		// The time fields are set to the default values.
		expected := Config{
			RyukReconnectionTimeout: 10 * time.Second,
			RyukConnectionTimeout:   time.Minute,
		}

		assert.Equal(t, expected, config)
	})

	t.Run("home/set/no-testcontainers-props-file/docker-host-env/set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir) // Windows support
		t.Setenv("DOCKER_HOST", tcpDockerHost33293)

		config, err := read()
		require.NoError(t, err)

		// The time fields are set to the default values,
		// and the config does not read DOCKER_HOST,
		// that's why it's empty
		expected := Config{
			RyukReconnectionTimeout: 10 * time.Second,
			RyukConnectionTimeout:   time.Minute,
		}

		assert.Equal(t, expected, config)
	})

	t.Run("home/set/no-testcontainers-props-file/testcontainers-ev/set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir) // Windows support
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", defaultHubPrefix)
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")
		t.Setenv("RYUK_VERBOSE", "true")
		t.Setenv("RYUK_RECONNECTION_TIMEOUT", "13s")
		t.Setenv("RYUK_CONNECTION_TIMEOUT", "12s")

		config, err := read()
		require.NoError(t, err)

		expected := Config{
			HubImageNamePrefix:      defaultHubPrefix,
			RyukDisabled:            true,
			RyukPrivileged:          true,
			RyukVerbose:             true,
			RyukReconnectionTimeout: 13 * time.Second,
			RyukConnectionTimeout:   12 * time.Second,
		}

		assert.Equal(t, expected, config)
	})

	t.Run("home/set/with-testcontainers-properties-file", func(t *testing.T) {
		defaultRyukConnectionTimeout := 60 * time.Second
		defaultRyukReconnectionTimeout := 10 * time.Second
		defaultCfg := Config{
			RyukConnectionTimeout:   defaultRyukConnectionTimeout,
			RyukReconnectionTimeout: defaultRyukReconnectionTimeout,
		}

		tests := []struct {
			name     string
			content  string
			env      map[string]string
			expected Config
			wantErr  bool
		}{
			{
				name:    "single-docker-host/with-spaces",
				content: "docker.host = " + tcpDockerHost33293,
				env:     map[string]string{},
				expected: Config{
					Host: tcpDockerHost33293,
				},
			},
			{
				name: "multiple-docker-hosts/last-one-wins",
				content: `docker.host = ` + tcpDockerHost33293 + `
	docker.host = ` + tcpDockerHost4711 + `
	`,
				env: map[string]string{},
				expected: Config{
					Host: tcpDockerHost4711,
				},
			},
			{
				name: "multiple-docker-hosts/last-one-wins/with-tls",
				content: `docker.host = ` + tcpDockerHost33293 + `
	docker.host = ` + tcpDockerHost4711 + `
	docker.host = ` + tcpDockerHost1234 + `
	docker.tls.verify = 1
	`,
				env: map[string]string{},
				expected: Config{
					Host:      tcpDockerHost1234,
					TLSVerify: 1,
				},
			},
			{
				name:     "properties/empty",
				content:  "",
				env:      map[string]string{},
				expected: Config{},
				wantErr:  false,
			},
			{
				name: "non-valid-properties-are-ignored",
				content: `foo = bar
	docker.host = ` + tcpDockerHost1234 + `
			`,
				env: map[string]string{},
				expected: Config{
					Host: tcpDockerHost1234,
				},
			},
			{
				name:    "single-docker-host/without-spaces",
				content: "docker.host=" + tcpDockerHost33293,
				env:     map[string]string{},
				expected: Config{
					Host: tcpDockerHost33293,
				},
			},
			{
				name:     "comments-are-ignored",
				content:  `#docker.host=` + tcpDockerHost33293,
				env:      map[string]string{},
				expected: defaultCfg,
				wantErr:  false,
			},
			{
				name: "multiple-docker-hosts/last-one-wins/with-tls/cert-path",
				content: `#docker.host = ` + tcpDockerHost33293 + `
	docker.host = ` + tcpDockerHost4711 + `
	docker.host = ` + tcpDockerHost1234 + `
	docker.cert.path=/tmp/certs`,
				env: map[string]string{},
				expected: Config{
					Host:     tcpDockerHost1234,
					CertPath: "/tmp/certs",
				},
			},
			{
				name:    "using-properties/ryuk-disabled",
				content: `ryuk.disabled=true`,
				env:     map[string]string{},
				expected: Config{
					RyukDisabled: true,
				},
			},
			{
				name:    "properties/ryuk-container-privileged",
				content: `ryuk.container.privileged=true`,
				env:     map[string]string{},
				expected: Config{
					RyukPrivileged: true,
				},
			},
			{
				name: "properties/ryuk-container-timeouts",
				content: `ryuk.connection.timeout=12s
	ryuk.reconnection.timeout=13s`,
				env: map[string]string{},
				expected: Config{
					RyukReconnectionTimeout: 13 * time.Second,
					RyukConnectionTimeout:   12 * time.Second,
				},
			},
			{
				name:    "env-vars/ryuk-container-timeouts",
				content: ``,
				env: map[string]string{
					"RYUK_RECONNECTION_TIMEOUT": "13s",
					"RYUK_CONNECTION_TIMEOUT":   "12s",
				},
				expected: Config{
					RyukReconnectionTimeout: 13 * time.Second,
					RyukConnectionTimeout:   12 * time.Second,
				},
			},
			{
				name: "env-vars/ryuk-container-timeouts/env-var-wins",
				content: `ryuk.connection.timeout=22s
	ryuk.reconnection.timeout=23s`,
				env: map[string]string{
					"RYUK_RECONNECTION_TIMEOUT": "13s",
					"RYUK_CONNECTION_TIMEOUT":   "12s",
				},
				expected: Config{
					RyukReconnectionTimeout: 13 * time.Second,
					RyukConnectionTimeout:   12 * time.Second,
				},
			},
			{
				name:    "properties/ryuk-verbose",
				content: `ryuk.verbose=true`,
				env:     map[string]string{},
				expected: Config{
					RyukVerbose: true,
				},
			},
			{
				name:    "env-vars/ryuk-disabled",
				content: ``,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				expected: Config{
					RyukDisabled: true,
				},
			},
			{
				name:    "env-vars/ryuk-container-privileged",
				content: ``,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				expected: Config{
					RyukPrivileged: true,
				},
			},
			{
				name:    "env-vars/properties/ryuk-disabled/env-var-wins-0",
				content: `ryuk.disabled=true`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				expected: Config{
					RyukDisabled: true,
				},
			},
			{
				name:    "env-vars/properties/ryuk-disabled/env-var-wins-1",
				content: `ryuk.disabled=false`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				expected: Config{
					RyukDisabled: true,
				},
			},
			{
				name:    "env-vars/properties/ryuk-disabled/env-var-wins-2",
				content: `ryuk.disabled=true`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "false",
				},
				expected: defaultCfg,
				wantErr:  false,
			},
			{
				name:    "env-vars/properties/ryuk-disabled/env-var-wins-3",
				content: `ryuk.disabled=false`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "false",
				},
				expected: defaultCfg,
				wantErr:  false,
			},
			{
				name:    "env-vars/properties/ryuk-verbose/env-var-wins-0",
				content: `ryuk.verbose=true`,
				env: map[string]string{
					"RYUK_VERBOSE": "true",
				},
				expected: Config{
					RyukVerbose: true,
				},
			},
			{
				name:    "env-vars/properties/ryuk-verbose/env-var-wins-1",
				content: `ryuk.verbose=false`,
				env: map[string]string{
					"RYUK_VERBOSE": "true",
				},
				expected: Config{
					RyukVerbose: true,
				},
			},
			{
				name:    "env-vars/properties/ryuk-verbose/env-var-wins-2",
				content: `ryuk.verbose=true`,
				env: map[string]string{
					"RYUK_VERBOSE": "false",
				},
				expected: defaultCfg,
				wantErr:  false,
			},
			{
				name:    "env-vars/properties/ryuk-verbose/env-var-wins-3",
				content: `ryuk.verbose=false`,
				env: map[string]string{
					"RYUK_VERBOSE": "false",
				},
				expected: defaultCfg,
				wantErr:  false,
			},
			{
				name:    "env-vars/properties/ryuk-container-privileged/env-var-wins-0",
				content: `ryuk.container.privileged=true`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				expected: Config{
					RyukPrivileged: true,
				},
			},
			{
				name:    "env-vars/properties/ryuk-container-privileged/env-var-wins-1",
				content: `ryuk.container.privileged=false`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				expected: Config{
					RyukPrivileged: true,
				},
			},
			{
				name:    "env-vars/properties/ryuk-container-privileged/env-var-wins-2",
				content: `ryuk.container.privileged=true`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "false",
				},
				expected: defaultCfg,
				wantErr:  false,
			},
			{
				name:    "env-vars/properties/ryuk-container-privileged/env-var-wins-3",
				content: `ryuk.container.privileged=false`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "false",
				},
				expected: defaultCfg,
				wantErr:  false,
			},
			{
				name: "properties/tls-verify/wrong-value",
				content: `ryuk.container.privileged=false
				docker.tls.verify = ERROR`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED":             "true",
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				expected: Config{
					RyukDisabled:   true,
					RyukPrivileged: true,
				},
				wantErr: true,
			},
			{
				name:    "env-vars/properties/ryuk-disabled/env-var-does-not-win/wrong-boolean",
				content: `ryuk.disabled=false`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "foo",
				},
				expected: defaultCfg,
				wantErr:  true,
			},
			{
				name:    "env-vars/properties/ryuk-container-privileged/env-var-does-not-win/wrong-boolean",
				content: `ryuk.container.privileged=false`,
				env: map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "foo",
				},
				expected: defaultCfg,
				wantErr:  true,
			},
			{
				name:    "properties/hub-image-name-prefix",
				content: `hub.image.name.prefix=` + defaultHubPrefix + `/props/`,
				env:     map[string]string{},
				expected: Config{
					HubImageNamePrefix: defaultHubPrefix + "/props/",
				},
			},
			{
				name:    "env-vars/hub-image-name-prefix",
				content: ``,
				env: map[string]string{
					"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX": defaultHubPrefix + "/env/",
				},
				expected: Config{
					HubImageNamePrefix: defaultHubPrefix + "/env/",
				},
			},
			{
				name:    "env-vars/properties/hub-image-name-prefix/env-var-wins",
				content: `hub.image.name.prefix=` + defaultHubPrefix + `/props/`,
				env: map[string]string{
					"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX": defaultHubPrefix + "/env/",
				},
				expected: Config{
					HubImageNamePrefix: defaultHubPrefix + "/env/",
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

				config, err := read()
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				// Merge the returned config, and the expected one, with the default config
				// to avoid setting all the fields in the expected config.
				// In the case of decoding errors in the properties file, the read config
				// needs to be merged with the default config to avoid setting the fields
				// that are not set in the properties file.
				err = mergo.Merge(&config, defaultCfg)
				require.NoError(t, err)

				err = mergo.Merge(&tt.expected, defaultCfg)
				require.NoError(t, err)

				require.Equal(t, tt.expected, config, "Configuration doesn't not match")
			})
		}
	})
}
