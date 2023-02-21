package testcontainers

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDockerConfig(t *testing.T) {
	t.Run("without env var retrieves default", func(t *testing.T) {
		cfg, err := getDockerConfig()
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 1, len(cfg.AuthConfigs))

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs["https://index.docker.io/v1/"]; !ok {
			t.Errorf("Expected to find https://index.docker.io/v1/ in auth configs")
		}
	})

	t.Run("with env var pointing to a non-existing file raises error", func(t *testing.T) {
		t.Setenv("DOCKER_CONFIG", filepath.Join("testresources", ".docker", "non-existing"))

		cfg, err := getDockerConfig()
		require.NotNil(t, err)
		require.Empty(t, cfg)
	})

	t.Run("with env var", func(t *testing.T) {
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
}
