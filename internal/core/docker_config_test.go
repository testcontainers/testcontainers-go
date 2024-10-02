package core

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuguy83/dockercfg"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/.docker/config.json
var dockerConfig string

func TestReadDockerConfig(t *testing.T) {
	expectedConfig := &dockercfg.Config{
		AuthConfigs: map[string]dockercfg.AuthConfig{
			IndexDockerIO:                 {},
			"https://example.com":         {},
			"https://my.private.registry": {},
		},
		CredentialsStore: "desktop",
	}
	t.Run("HOME/valid", func(t *testing.T) {
		testDockerConfigHome(t, "testdata")

		cfg, err := ReadDockerConfig()
		require.NoError(t, err)
		require.Equal(t, expectedConfig, cfg)
	})

	t.Run("HOME/not-found", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")

		cfg, err := ReadDockerConfig()
		require.ErrorIs(t, err, os.ErrNotExist)
		require.Nil(t, cfg)
	})

	t.Run("HOME/invalid-config", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "invalid-config")

		cfg, err := ReadDockerConfig()
		require.ErrorContains(t, err, "json: cannot unmarshal array")
		require.Nil(t, cfg)
	})

	t.Run("DOCKER_AUTH_CONFIG/valid", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")
		t.Setenv("DOCKER_AUTH_CONFIG", dockerConfig)

		cfg, err := ReadDockerConfig()
		require.NoError(t, err)
		require.Equal(t, expectedConfig, cfg)
	})

	t.Run("DOCKER_AUTH_CONFIG/invalid-config", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")
		t.Setenv("DOCKER_AUTH_CONFIG", `{"auths": []}`)

		cfg, err := ReadDockerConfig()
		require.ErrorContains(t, err, "json: cannot unmarshal array")
		require.Nil(t, cfg)
	})

	t.Run("DOCKER_CONFIG/valid", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")
		t.Setenv("DOCKER_CONFIG", filepath.Join("testdata", ".docker"))

		cfg, err := ReadDockerConfig()
		require.NoError(t, err)
		require.Equal(t, expectedConfig, cfg)
	})

	t.Run("DOCKER_CONFIG/invalid-config", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")
		t.Setenv("DOCKER_CONFIG", filepath.Join("testdata", "invalid-config", ".docker"))

		cfg, err := ReadDockerConfig()
		require.ErrorContains(t, err, "json: cannot unmarshal array")
		require.Nil(t, cfg)
	})
}

// testDockerConfigHome sets the user's home directory to the given path
// and unsets the DOCKER_CONFIG and DOCKER_AUTH_CONFIG environment variables.
func testDockerConfigHome(t *testing.T, dirs ...string) {
	t.Helper()

	dir := filepath.Join(dirs...)
	t.Setenv("DOCKER_AUTH_CONFIG", "")
	t.Setenv("DOCKER_CONFIG", "")
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir) // Windows
}
