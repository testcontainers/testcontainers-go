package testcontainers

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
)

func TestReadTCPropsFile(t *testing.T) {
	t.Run("HOME is not set", func(t *testing.T) {
		env.Patch(t, "HOME", "")

		config := configureTC()

		assert.Empty(t, config, "TC props file should not exist")
	})

	t.Run("HOME is not set - TESTCONTAINERS_ env is set", func(t *testing.T) {
		env.Patch(t, "HOME", "")
		env.Patch(t, "TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")

		config := configureTC()

		expected := TestcontainersConfig{}
		expected.RyukPrivileged = true

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file", func(t *testing.T) {
		tmpDir := fs.NewDir(t, os.TempDir())
		env.Patch(t, "HOME", tmpDir.Path())

		config := configureTC()

		assert.Empty(t, config, "TC props file should not exist")
	})

	t.Run("HOME does not contain TC props file - TESTCONTAINERS_ env is set", func(t *testing.T) {
		tmpDir := fs.NewDir(t, os.TempDir())
		env.Patch(t, "HOME", tmpDir.Path())
		env.Patch(t, "TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")

		config := configureTC()
		expected := TestcontainersConfig{}
		expected.RyukPrivileged = true

		assert.Equal(t, expected, config)
	})

	t.Run("HOME contains TC properties file", func(t *testing.T) {
		tests := []struct {
			content  string
			env      map[string]string
			expected TestcontainersConfig
		}{
			{
				"docker.host = tcp://127.0.0.1:33293",
				map[string]string{},
				TestcontainersConfig{
					Host:      "tcp://127.0.0.1:33293",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"docker.host = tcp://127.0.0.1:33293",
				map[string]string{},
				TestcontainersConfig{
					Host:      "tcp://127.0.0.1:33293",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
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
				"",
				map[string]string{},
				TestcontainersConfig{
					Host:      "",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
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
				"docker.host=tcp://127.0.0.1:33293",
				map[string]string{},
				TestcontainersConfig{
					Host:      "tcp://127.0.0.1:33293",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				`#docker.host=tcp://127.0.0.1:33293`,
				map[string]string{},
				TestcontainersConfig{
					Host:      "",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
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
				`ryuk.container.privileged=false
				docker.tls.verify = ERROR`,
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
		for i, tt := range tests {
			t.Run(fmt.Sprintf("[%d]", i), func(t *testing.T) {
				tmpDir := fs.NewDir(t, os.TempDir())
				env.Patch(t, "HOME", tmpDir.Path())
				for k, v := range tt.env {
					env.Patch(t, k, v)
				}
				if err := ioutil.WriteFile(tmpDir.Join(".testcontainers.properties"), []byte(tt.content), 0o600); err != nil {
					t.Errorf("Failed to create the file: %v", err)
					return
				}

				config := configureTC()

				assert.Equal(t, tt.expected, config, "Configuration doesn't not match")

			})
		}
	})
}
