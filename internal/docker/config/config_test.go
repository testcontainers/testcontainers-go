package config

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/.docker/config.json
var dockerConfig string

func TestReadDockerConfig(t *testing.T) {
	var expectedConfig Config
	err := json.Unmarshal([]byte(dockerConfig), &expectedConfig)
	require.NoError(t, err)

	setupDockerConfigs(t, "")

	t.Run("HOME", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			setupHome(t, "testdata")

			cfg, err := Load()
			require.NoError(t, err)
			require.Equal(t, expectedConfig, cfg)
		})

		t.Run("not-found", func(t *testing.T) {
			setupHome(t, "testdata", "not-found")

			cfg, err := Load()
			require.ErrorIs(t, err, os.ErrNotExist)
			require.Empty(t, cfg)
		})

		t.Run("invalid-config", func(t *testing.T) {
			setupHome(t, "testdata", "invalid-config")

			cfg, err := Load()
			require.ErrorContains(t, err, "json: cannot unmarshal array")
			require.Empty(t, cfg)
		})
	})

	t.Run("DOCKER_AUTH_CONFIG", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			setupHome(t, "testdata", "not-found")
			t.Setenv("DOCKER_AUTH_CONFIG", dockerConfig)

			cfg, err := Load()
			require.NoError(t, err)
			require.Equal(t, expectedConfig, cfg)
		})

		t.Run("invalid-config", func(t *testing.T) {
			setupHome(t, "testdata", "not-found")
			t.Setenv("DOCKER_AUTH_CONFIG", `{"auths": []}`)

			cfg, err := Load()
			require.ErrorContains(t, err, "json: cannot unmarshal array")
			require.Empty(t, cfg)
		})
	})

	t.Run(EnvOverrideDir, func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			setupHome(t, "testdata", "not-found")
			t.Setenv(EnvOverrideDir, filepath.Join("testdata", ".docker"))

			cfg, err := Load()
			require.NoError(t, err)
			require.Equal(t, expectedConfig, cfg)
		})

		t.Run("invalid-config", func(t *testing.T) {
			setupHome(t, "testdata", "not-found")
			t.Setenv(EnvOverrideDir, filepath.Join("testdata", "invalid-config", ".docker"))

			cfg, err := Load()
			require.ErrorContains(t, err, "json: cannot unmarshal array")
			require.Empty(t, cfg)
		})
	})
}

func TestDir(t *testing.T) {
	setupDockerConfigs(t, "")

	t.Run("HOME", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			tmpDir := t.TempDir()
			setupHome(t, tmpDir)

			dir, err := Dir()
			require.NoError(t, err)
			require.Equal(t, filepath.Join(tmpDir, ".docker"), dir)
		})
	})

	t.Run(EnvOverrideDir, func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			tmpDir := t.TempDir()
			setupDockerConfigs(t, tmpDir)

			dir, err := Dir()
			require.NoError(t, err)
			require.Equal(t, tmpDir, dir)
		})
	})
}

// setupHome sets the user's home directory to the given path
// and unsets the DOCKER_CONFIG and DOCKER_AUTH_CONFIG environment variables.
func setupHome(t *testing.T, dirs ...string) {
	t.Helper()

	dir := filepath.Join(dirs...)
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir) // Windows
}

// setupHome sets the user's home directory to the given path
// and unsets the DOCKER_CONFIG and DOCKER_AUTH_CONFIG environment variables.
func setupDockerConfigs(t *testing.T, dirs ...string) {
	t.Helper()

	dir := filepath.Join(dirs...)
	t.Setenv("DOCKER_AUTH_CONFIG", dir)
	t.Setenv(EnvOverrideDir, dir)
}
