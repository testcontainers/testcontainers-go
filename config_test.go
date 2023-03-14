package testcontainers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoReadTCConfig(t *testing.T) {
	t.Run("HOME is not set", func(t *testing.T) {
		t.Setenv("HOME", "")

		config := doReadConfig()

		assert.Empty(t, config, "TC props file should not exist")
	})

	t.Run("HOME is not set - TESTCONTAINERS_ env is set", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")

		config := doReadConfig()

		expected := TestcontainersConfig{}
		expected.RyukDisabled = true
		expected.RyukPrivileged = true

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		config := doReadConfig()

		assert.Empty(t, config, "TC props file should not exist")
	})

	t.Run("HOME does not contain TC props file - TESTCONTAINERS_ env is set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")

		config := doReadConfig()
		expected := TestcontainersConfig{}
		expected.RyukDisabled = true
		expected.RyukPrivileged = true

		assert.Equal(t, expected, config)
	})

	t.Run("HOME contains TC properties file", func(t *testing.T) {
		tests := []struct {
			name     string
			content  string
			env      map[string]string
			expected TestcontainersConfig
		}{
			{
				"Single Docker host with spaces",
				"docker.host = tcp://127.0.0.1:33293",
				map[string]string{},
				TestcontainersConfig{
					Host:      "tcp://127.0.0.1:33293",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Multiple docker host entries, last one wins",
				`docker.host = tcp://127.0.0.1:33293
	docker.host = tcp://127.0.0.1:4711
	`,
				map[string]string{},
				TestcontainersConfig{
					Host:      "tcp://127.0.0.1:4711",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Multiple docker host entries, last one wins, with TLS",
				`docker.host = tcp://127.0.0.1:33293
	docker.host = tcp://127.0.0.1:4711
	docker.host = tcp://127.0.0.1:1234
	docker.tls.verify = 1
	`,
				map[string]string{},
				TestcontainersConfig{
					Host:      "tcp://127.0.0.1:1234",
					TLSVerify: 1,
					CertPath:  "",
				},
			},
			{
				"Empty file",
				"",
				map[string]string{},
				TestcontainersConfig{
					Host:      "",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Non-valid properties are ignored",
				`foo = bar
	docker.host = tcp://127.0.0.1:1234
			`,
				map[string]string{},
				TestcontainersConfig{
					Host:      "tcp://127.0.0.1:1234",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Single Docker host without spaces",
				"docker.host=tcp://127.0.0.1:33293",
				map[string]string{},
				TestcontainersConfig{
					Host:      "tcp://127.0.0.1:33293",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Comments are ignored",
				`#docker.host=tcp://127.0.0.1:33293`,
				map[string]string{},
				TestcontainersConfig{
					Host:      "",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"Multiple docker host entries, last one wins, with TLS and cert path",
				`#docker.host = tcp://127.0.0.1:33293
	docker.host = tcp://127.0.0.1:4711
	docker.host = tcp://127.0.0.1:1234
	docker.cert.path=/tmp/certs`,
				map[string]string{},
				TestcontainersConfig{
					Host:      "tcp://127.0.0.1:1234",
					TLSVerify: 0,
					CertPath:  "/tmp/certs",
				},
			},
			{
				"With Ryuk disabled using properties",
				`ryuk.disabled=true`,
				map[string]string{},
				TestcontainersConfig{
					Host:         "",
					TLSVerify:    0,
					CertPath:     "",
					RyukDisabled: true,
				},
			},
			{
				"With Ryuk container privileged using properties",
				`ryuk.container.privileged=true`,
				map[string]string{},
				TestcontainersConfig{
					Host:           "",
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
				TestcontainersConfig{
					Host:         "",
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
				TestcontainersConfig{
					Host:           "",
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
				TestcontainersConfig{
					Host:         "",
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
				TestcontainersConfig{
					Host:         "",
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
				TestcontainersConfig{
					Host:         "",
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
				TestcontainersConfig{
					Host:         "",
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
				TestcontainersConfig{
					Host:           "",
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
				TestcontainersConfig{
					Host:           "",
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
				TestcontainersConfig{
					Host:           "",
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
				TestcontainersConfig{
					Host:           "",
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
				TestcontainersConfig{
					Host:           "",
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
				TestcontainersConfig{
					Host:         "",
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
				TestcontainersConfig{
					Host:           "",
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

				config := doReadConfig()

				assert.Equal(t, tt.expected, config, "Configuration doesn't not match")

			})
		}
	})
}
