package auth

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cpuguy83/dockercfg"
	"github.com/docker/docker/api/types/registry"
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
		setAuthConfig(t, exampleAuth, "", "")
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
		username, password := "gopher", "secret"
		creds := setAuthConfig(t, exampleAuth, username, password)

		registry, cfg, err := ForDockerImage(context.Background(), exampleAuth+"/my/image:latest")
		require.NoError(t, err)
		require.NotEmpty(t, cfg)

		assert.Equal(t, exampleAuth, registry)
		assert.Equal(t, username, cfg.Username)
		assert.Equal(t, password, cfg.Password)
		assert.Equal(t, creds, cfg.Auth)
	})

	t.Run("match registry authentication by host", func(t *testing.T) {
		imageReg := "example-auth.com"
		imagePath := "/my/image:latest"
		base64 := setAuthConfig(t, exampleAuth, "gopher", "secret")

		registry, cfg, err := ForDockerImage(context.Background(), imageReg+imagePath)
		require.NoError(t, err)
		require.NotEmpty(t, cfg)

		assert.Equal(t, imageReg, registry)
		assert.Equal(t, "gopher", cfg.Username)
		assert.Equal(t, "secret", cfg.Password)
		assert.Equal(t, base64, cfg.Auth)
	})

	t.Run("fail to match registry authentication due to invalid host", func(t *testing.T) {
		imageReg := "example-auth.com"
		imagePath := "/my/image:latest"
		invalidRegistryURL := "://invalid-host"
		setAuthConfig(t, invalidRegistryURL, "gopher", "secret")

		registry, cfg, err := ForDockerImage(context.Background(), imageReg+imagePath)
		require.ErrorIs(t, err, dockercfg.ErrCredentialsNotFound)
		require.Empty(t, cfg)

		assert.Equal(t, imageReg, registry)
	})

	t.Run("fail to match registry authentication by host with empty URL scheme creds and missing default", func(t *testing.T) {
		origDefaultRegistryFn := defaultRegistryFn
		t.Cleanup(func() {
			defaultRegistryFn = origDefaultRegistryFn
		})
		defaultRegistryFn = func(ctx context.Context) string {
			return ""
		}

		imageReg := ""
		imagePath := "image:latest"
		setAuthConfig(t, "example-auth.com", "gopher", "secret")

		registry, cfg, err := ForDockerImage(context.Background(), imageReg+imagePath)
		require.ErrorIs(t, err, dockercfg.ErrCredentialsNotFound)
		require.Empty(t, cfg)

		assert.Equal(t, imageReg, registry)
	})
}

//go:embed testdata/.docker/config.json
var dockerConfig string

func TestGetDockerConfigs(t *testing.T) {
	t.Run("file", func(t *testing.T) {
		got, err := GetDockerConfigs()
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("env", func(t *testing.T) {
		t.Setenv("DOCKER_AUTH_CONFIG", dockerConfig)

		got, err := GetDockerConfigs()
		require.NoError(t, err)

		// We can only check the keys as the values are not deterministic.
		expected := map[string]registry.AuthConfig{
			"https://index.docker.io/v1/": {},
			"https://example.com":         {},
			"https://my.private.registry": {},
		}
		for k := range got {
			got[k] = registry.AuthConfig{}
		}
		require.Equal(t, expected, got)
	})
}

// setAuthConfig sets the DOCKER_AUTH_CONFIG environment variable with
// authentication for with the given host, username and password.
// It returns the base64 encoded credentials.
func setAuthConfig(t *testing.T, host, username, password string) string {
	t.Helper()

	var creds string
	if username != "" || password != "" {
		creds = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	}

	auth := fmt.Sprintf(`{
		"auths": {
			%q: {
				"username": %q,
				"password": %q,
				"auth": %q
			}
		},
		"credsStore": "desktop"
	}`,
		host,
		username,
		password,
		creds,
	)
	t.Setenv("DOCKER_AUTH_CONFIG", auth)

	return creds
}
