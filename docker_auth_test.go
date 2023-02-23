package testcontainers

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDockerConfig(t *testing.T) {
	t.Run("without DOCKER_CONFIG env var retrieves default", func(t *testing.T) {
		cfg, err := getDockerConfig()
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 1, len(cfg.AuthConfigs))

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs["https://index.docker.io/v1/"]; !ok {
			t.Errorf("Expected to find https://index.docker.io/v1/ in auth configs")
		}
	})

	t.Run("with DOCKER_CONFIG env var pointing to a non-existing file raises error", func(t *testing.T) {
		t.Setenv("DOCKER_CONFIG", filepath.Join("testresources", ".docker", "non-existing"))

		cfg, err := getDockerConfig()
		require.NotNil(t, err)
		require.Empty(t, cfg)
	})

	t.Run("with DOCKER_CONFIG env var", func(t *testing.T) {
		t.Setenv("DOCKER_CONFIG", filepath.Join("testresources", ".docker"))

		cfg, err := getDockerConfig()
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 3, len(cfg.AuthConfigs))

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs["https://index.docker.io/v1/"]; !ok {
			t.Errorf("Expected to find https://index.docker.io/v1/ in auth configs")
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
					"https://example-auth.com": {}
			},
			"credsStore": "desktop"
		}`)
		t.Setenv("DOCKER_CONFIG", filepath.Join("testresources", ".docker"))

		cfg, err := getDockerConfig()
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 1, len(cfg.AuthConfigs))

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs["https://index.docker.io/v1/"]; ok {
			t.Errorf("Not expected to find https://index.docker.io/v1/ in auth configs")
		}
		if _, ok := authCfgs["https://example-auth.com"]; !ok {
			t.Errorf("Expected to find https://example-auth.com in auth configs")
		}
	})

	t.Run("retrieve auth with DOCKER_AUTH_CONFIG env var", func(t *testing.T) {
		base64 := "Z29waGVyOnNlY3JldA==" // gopher:secret

		t.Setenv("DOCKER_AUTH_CONFIG", `{
			"auths": {
					"https://example-auth.com": { "username": "gopher", "password": "secret", "auth": "`+base64+`" }
			},
			"credsStore": "desktop"
		}`)

		cfg, err := AuthFromDockerConfig("https://example-auth.com")
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "gopher", cfg.Username)
		assert.Equal(t, "secret", cfg.Password)
		assert.Equal(t, base64, cfg.Auth)
	})
}
