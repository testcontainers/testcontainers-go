package testcontainers

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleAuth = "https://example-auth.com"
const indexDockerIO = "https://index.docker.io/v1/"

var testDockerConfigDirPath = filepath.Join("testresources", ".docker")

func TestGetDockerConfig(t *testing.T) {
	t.Run("without DOCKER_CONFIG env var retrieves default", func(t *testing.T) {
		cfg, err := getDockerConfig()
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 1, len(cfg.AuthConfigs))

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs[indexDockerIO]; !ok {
			t.Errorf("Expected to find %s in auth configs", indexDockerIO)
		}
	})

	t.Run("with DOCKER_CONFIG env var pointing to a non-existing file raises error", func(t *testing.T) {
		t.Setenv("DOCKER_CONFIG", filepath.Join(testDockerConfigDirPath, "non-existing"))

		cfg, err := getDockerConfig()
		require.NotNil(t, err)
		require.Empty(t, cfg)
	})

	t.Run("with DOCKER_CONFIG env var", func(t *testing.T) {
		t.Setenv("DOCKER_CONFIG", testDockerConfigDirPath)

		cfg, err := getDockerConfig()
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 3, len(cfg.AuthConfigs))

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs[indexDockerIO]; !ok {
			t.Errorf("Expected to find %s in auth configs", indexDockerIO)
		}
		if _, ok := authCfgs["https://example.com"]; !ok {
			t.Errorf("Expected to find https://example.com in auth configs")
		}
		if _, ok := authCfgs["https://my.private.registry"]; !ok {
			t.Errorf("Expected to find https://my.private.registry in auth configs")
		}
	})

	t.Run("DOCKER_AUTH_CONFIG env var takes precedence", func(t *testing.T) {
		t.Setenv("DOCKER_AUTH_CONFIG", `{
			"auths": {
					"`+exampleAuth+`": {}
			},
			"credsStore": "desktop"
		}`)
		t.Setenv("DOCKER_CONFIG", testDockerConfigDirPath)

		cfg, err := getDockerConfig()
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 1, len(cfg.AuthConfigs))

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs[indexDockerIO]; ok {
			t.Errorf("Not expected to find %s in auth configs", indexDockerIO)
		}
		if _, ok := authCfgs[exampleAuth]; !ok {
			t.Errorf("Expected to find %s in auth configs", exampleAuth)
		}
	})

	t.Run("retrieve auth with DOCKER_AUTH_CONFIG env var", func(t *testing.T) {
		base64 := "Z29waGVyOnNlY3JldA==" // gopher:secret

		t.Setenv("DOCKER_AUTH_CONFIG", `{
			"auths": {
					"`+exampleAuth+`": { "username": "gopher", "password": "secret", "auth": "`+base64+`" }
			},
			"credsStore": "desktop"
		}`)

		cfg, err := AuthFromDockerConfig(exampleAuth)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "gopher", cfg.Username)
		assert.Equal(t, "secret", cfg.Password)
		assert.Equal(t, base64, cfg.Auth)
	})
}
