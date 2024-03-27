// This test is testing very internal logic that should not be exported away from this package. We'll
// leave it in the main testcontainers package. Do not use for user facing examples.
package testcontainers

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/cpuguy83/dockercfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

const exampleAuth = "https://example-auth.com"

var testDockerConfigDirPath = filepath.Join("testdata", ".docker")

var indexDockerIO = core.IndexDockerIO

func TestGetDockerConfig(t *testing.T) {
	const expectedErrorMessage = "Expected to find %s in auth configs"

	// Verify that the default docker config file exists before any test in this suite runs.
	// Then, we can safely run the tests that rely on it.
	defaultCfg, err := dockercfg.LoadDefaultConfig()
	require.NoError(t, err)
	require.NotEmpty(t, defaultCfg)

	t.Run("without DOCKER_CONFIG env var retrieves default", func(t *testing.T) {
		t.Setenv("DOCKER_CONFIG", "")

		cfg, err := getDockerConfig()
		require.NoError(t, err)
		require.NotEmpty(t, cfg)

		assert.Equal(t, defaultCfg, cfg)
	})

	t.Run("with DOCKER_CONFIG env var pointing to a non-existing file raises error", func(t *testing.T) {
		t.Setenv("DOCKER_CONFIG", filepath.Join(testDockerConfigDirPath, "non-existing"))

		cfg, err := getDockerConfig()
		require.Error(t, err)
		require.Empty(t, cfg)
	})

	t.Run("with DOCKER_CONFIG env var", func(t *testing.T) {
		t.Setenv("DOCKER_CONFIG", testDockerConfigDirPath)

		cfg, err := getDockerConfig()
		require.NoError(t, err)
		require.NotEmpty(t, cfg)

		assert.Len(t, cfg.AuthConfigs, 3)

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs[indexDockerIO]; !ok {
			t.Errorf(expectedErrorMessage, indexDockerIO)
		}
		if _, ok := authCfgs["https://example.com"]; !ok {
			t.Errorf(expectedErrorMessage, "https://example.com")
		}
		if _, ok := authCfgs["https://my.private.registry"]; !ok {
			t.Errorf(expectedErrorMessage, "https://my.private.registry")
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
		require.NoError(t, err)
		require.NotEmpty(t, cfg)

		assert.Len(t, cfg.AuthConfigs, 1)

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs[indexDockerIO]; ok {
			t.Errorf("Not expected to find %s in auth configs", indexDockerIO)
		}
		if _, ok := authCfgs[exampleAuth]; !ok {
			t.Errorf(expectedErrorMessage, exampleAuth)
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

		registry, cfg, err := DockerImageAuth(context.Background(), exampleAuth+"/my/image:latest")
		require.NoError(t, err)
		require.NotEmpty(t, cfg)

		assert.Equal(t, exampleAuth, registry)
		assert.Equal(t, "gopher", cfg.Username)
		assert.Equal(t, "secret", cfg.Password)
		assert.Equal(t, base64, cfg.Auth)
	})

	t.Run("match registry authentication by host", func(t *testing.T) {
		base64 := "Z29waGVyOnNlY3JldA==" // gopher:secret
		imageReg := "example-auth.com"
		imagePath := "/my/image:latest"

		t.Setenv("DOCKER_AUTH_CONFIG", `{
			"auths": {
					"`+exampleAuth+`": { "username": "gopher", "password": "secret", "auth": "`+base64+`" }
			},
			"credsStore": "desktop"
		}`)

		registry, cfg, err := DockerImageAuth(context.Background(), imageReg+imagePath)
		require.NoError(t, err)
		require.NotEmpty(t, cfg)

		assert.Equal(t, imageReg, registry)
		assert.Equal(t, "gopher", cfg.Username)
		assert.Equal(t, "secret", cfg.Password)
		assert.Equal(t, base64, cfg.Auth)
	})

	t.Run("fail to match registry authentication due to invalid host", func(t *testing.T) {
		base64 := "Z29waGVyOnNlY3JldA==" // gopher:secret
		imageReg := "example-auth.com"
		imagePath := "/my/image:latest"
		invalidRegistryURL := "://invalid-host"

		t.Setenv("DOCKER_AUTH_CONFIG", `{
			"auths": {
					"`+invalidRegistryURL+`": { "username": "gopher", "password": "secret", "auth": "`+base64+`" }
			},
			"credsStore": "desktop"
		}`)

		registry, cfg, err := DockerImageAuth(context.Background(), imageReg+imagePath)
		require.Equal(t, err, dockercfg.ErrCredentialsNotFound)
		require.Empty(t, cfg)

		assert.Equal(t, imageReg, registry)
	})
}
