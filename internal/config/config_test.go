package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

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
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "")
	t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "")
}

func TestReadConfig(t *testing.T) {
	resetTestEnv(t)

	t.Run("Config is read just once", func(t *testing.T) {
		t.Setenv("HOME", "")
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

	t.Run("HOME is not set", func(t *testing.T) {
		t.Setenv("HOME", "")

		config := read()

		expected := Config{}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME is not set - TESTCONTAINERS_ env is set", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")

		config := read()

		expected := Config{
			RyukDisabled:   true,
			RyukPrivileged: true,
			Host:           "", // docker socket is empty at the properties file
		}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		config := read()

		expected := Config{}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file - DOCKER_HOST env is set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("DOCKER_HOST", tcpDockerHost33293)

		config := read()
		expected := Config{} // the config does not read DOCKER_HOST, that's why it's empty

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file - TESTCONTAINERS_ env is set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")

		config := read()
		expected := Config{
			RyukDisabled:   true,
			RyukPrivileged: true,
		}

		assert.Equal(t, expected, config)
	})

	t.Run("HOME contains TC properties file", func(t *testing.T) {
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
					Host:      tcpDockerHost33293,
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Multiple docker host entries, last one wins",
				`docker.host = ` + tcpDockerHost33293 + `
	docker.host = ` + tcpDockerHost4711 + `
	`,
				map[string]string{},
				Config{
					Host:      tcpDockerHost4711,
					TLSVerify: 0,
					CertPath:  "",
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
					Host:      tcpDockerHost1234,
					TLSVerify: 1,
					CertPath:  "",
				},
			},
			{
				"Empty file",
				"",
				map[string]string{},
				Config{
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Non-valid properties are ignored",
				`foo = bar
	docker.host = ` + tcpDockerHost1234 + `
			`,
				map[string]string{},
				Config{
					Host:      tcpDockerHost1234,
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Single Docker host without spaces",
				"docker.host=" + tcpDockerHost33293,
				map[string]string{},
				Config{
					Host:      tcpDockerHost33293,
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Comments are ignored",
				`#docker.host=` + tcpDockerHost33293,
				map[string]string{},
				Config{
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Multiple docker host entries, last one wins, with TLS and cert path",
				`#docker.host = ` + tcpDockerHost33293 + `
	docker.host = ` + tcpDockerHost4711 + `
	docker.host = ` + tcpDockerHost1234 + `
	docker.cert.path=/tmp/certs`,
				map[string]string{},
				Config{
					Host:      tcpDockerHost1234,
					TLSVerify: 0,
					CertPath:  "/tmp/certs",
				},
			},
			{
				"With Ryuk disabled using properties",
				`ryuk.disabled=true`,
				map[string]string{},
				Config{
					TLSVerify:    0,
					CertPath:     "",
					RyukDisabled: true,
				},
			},
			{
				"With Ryuk container privileged using properties",
				`ryuk.container.privileged=true`,
				map[string]string{},
				Config{
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: true,
				},
			},
			{
				"With Ryuk disabled using an env var",
				``,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				Config{
					TLSVerify:    0,
					CertPath:     "",
					RyukDisabled: true,
				},
			},
			{
				"With Ryuk container privileged using an env var",
				``,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				Config{
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: true,
				},
			},
			{
				"With Ryuk disabled using an env var and properties. Env var wins (0)",
				`ryuk.disabled=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				Config{
					TLSVerify:    0,
					CertPath:     "",
					RyukDisabled: true,
				},
			},
			{
				"With Ryuk disabled using an env var and properties. Env var wins (1)",
				`ryuk.disabled=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "true",
				},
				Config{
					TLSVerify:    0,
					CertPath:     "",
					RyukDisabled: true,
				},
			},
			{
				"With Ryuk disabled using an env var and properties. Env var wins (2)",
				`ryuk.disabled=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "false",
				},
				Config{
					TLSVerify:    0,
					CertPath:     "",
					RyukDisabled: false,
				},
			},
			{
				"With Ryuk disabled using an env var and properties. Env var wins (3)",
				`ryuk.disabled=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_DISABLED": "false",
				},
				Config{
					TLSVerify:    0,
					CertPath:     "",
					RyukDisabled: false,
				},
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var wins (0)",
				`ryuk.container.privileged=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				Config{
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: true,
				},
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var wins (1)",
				`ryuk.container.privileged=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				Config{
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: true,
				},
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var wins (2)",
				`ryuk.container.privileged=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "false",
				},
				Config{
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: false,
				},
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var wins (3)",
				`ryuk.container.privileged=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "false",
				},
				Config{
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: false,
				},
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
					TLSVerify:      0,
					CertPath:       "",
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
				Config{
					TLSVerify:    0,
					CertPath:     "",
					RyukDisabled: false,
				},
			},
			{
				"With Ryuk container privileged using an env var and properties. Env var does not win because it's not a boolean value",
				`ryuk.container.privileged=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "foo",
				},
				Config{
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: false,
				},
			},
		}
		for _, tt := range tests {
			t.Run(fmt.Sprintf(tt.name), func(t *testing.T) {
				tmpDir := t.TempDir()
				t.Setenv("HOME", tmpDir)
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
