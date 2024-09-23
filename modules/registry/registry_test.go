package registry_test

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/cpuguy83/dockercfg"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRegistry_unauthenticated(t *testing.T) {
	ctx := context.Background()
	ctr, err := registry.Run(ctx, registry.DefaultImage)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	httpAddress, err := ctr.Address(ctx)
	require.NoError(t, err)

	resp, err := http.Get(httpAddress + "/v2/_catalog")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRunContainer_authenticated(t *testing.T) {
	ctx := context.Background()
	registryContainer, err := registry.Run(
		ctx,
		registry.DefaultImage,
		registry.WithHtpasswdFile(filepath.Join("testdata", "auth", "htpasswd")),
		registry.WithData(filepath.Join("testdata", "data")),
	)
	testcontainers.CleanupContainer(t, registryContainer)
	require.NoError(t, err)

	// httpAddress {
	httpAddress, err := registryContainer.Address(ctx)
	// }
	require.NoError(t, err)

	registryHost, err := registryContainer.HostAddress(ctx)
	require.NoError(t, err)

	t.Run("HTTP connection without basic auth fails", func(tt *testing.T) {
		httpCli := http.Client{}
		req, err := http.NewRequest(http.MethodGet, httpAddress+"/v2/_catalog", nil)
		require.NoError(t, err)

		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("HTTP connection with incorrect basic auth fails", func(tt *testing.T) {
		httpCli := http.Client{}
		req, err := http.NewRequest(http.MethodGet, httpAddress+"/v2/_catalog", nil)
		require.NoError(t, err)

		req.SetBasicAuth("foo", "bar")

		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("HTTP connection with basic auth succeeds", func(tt *testing.T) {
		httpCli := http.Client{}
		req, err := http.NewRequest(http.MethodGet, httpAddress+"/v2/_catalog", nil)
		require.NoError(t, err)

		req.SetBasicAuth("testuser", "testpassword")

		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("build images with wrong credentials fails", func(tt *testing.T) {
		setAuthConfig(tt, registryHost, "foo", "bar")

		redisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context: filepath.Join("testdata", "redis"),
					BuildArgs: map[string]*string{
						"REGISTRY_HOST": &registryHost,
					},
				},
				AlwaysPullImage: true, // make sure the authentication takes place
				ExposedPorts:    []string{"6379/tcp"},
				WaitingFor:      wait.ForLog("Ready to accept connections"),
			},
			Started: true,
		})
		testcontainers.CleanupContainer(tt, redisC)
		require.Error(tt, err)
		require.Contains(tt, err.Error(), "unauthorized: authentication required")
	})

	t.Run("build image with valid credentials", func(tt *testing.T) {
		setAuthConfig(tt, registryHost, "testuser", "testpassword")

		// build a custom redis image from the private registry,
		// using RegistryName of the container as the registry.
		// The container should start because the authentication
		// is correct.

		redisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context: filepath.Join("testdata", "redis"),
					BuildArgs: map[string]*string{
						"REGISTRY_HOST": &registryHost,
					},
				},
				AlwaysPullImage: true, // make sure the authentication takes place
				ExposedPorts:    []string{"6379/tcp"},
				WaitingFor:      wait.ForLog("Ready to accept connections"),
			},
			Started: true,
		})
		testcontainers.CleanupContainer(tt, redisC)
		require.NoError(tt, err)

		state, err := redisC.State(context.Background())
		require.NoError(tt, err)
		require.True(tt, state.Running, "expected redis container to be running, but it is not")
	})
}

func TestRunContainer_authenticated_withCredentials(t *testing.T) {
	ctx := context.Background()
	// htpasswdString {
	registryContainer, err := registry.Run(
		ctx,
		registry.DefaultImage,
		registry.WithHtpasswd("testuser:$2y$05$tTymaYlWwJOqie.bcSUUN.I.kxmo1m5TLzYQ4/ejJ46UMXGtq78EO"),
	)
	// }
	testcontainers.CleanupContainer(t, registryContainer)
	require.NoError(t, err)

	httpAddress, err := registryContainer.Address(ctx)
	require.NoError(t, err)

	httpCli := http.Client{}
	req, err := http.NewRequest(http.MethodGet, httpAddress+"/v2/_catalog", nil)
	require.NoError(t, err)

	req.SetBasicAuth("testuser", "testpassword")

	resp, err := httpCli.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRunContainer_wrongData(t *testing.T) {
	ctx := context.Background()
	registryContainer, err := registry.Run(
		ctx,
		registry.DefaultImage,
		registry.WithHtpasswdFile(filepath.Join("testdata", "auth", "htpasswd")),
		registry.WithData(filepath.Join("testdata", "wrongdata")),
	)
	testcontainers.CleanupContainer(t, registryContainer)
	require.NoError(t, err)

	registryHost, err := registryContainer.HostAddress(ctx)
	require.NoError(t, err)

	setAuthConfig(t, registryHost, "testuser", "testpassword")

	// build a custom redis image from the private registry,
	// using RegistryName of the container as the registry.
	// The container won't be able to start because the data
	// directory is wrong.

	redisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context: filepath.Join("testdata", "redis"),
				BuildArgs: map[string]*string{
					"REGISTRY_HOST": &registryHost,
				},
			},
			AlwaysPullImage: true, // make sure the authentication takes place
			ExposedPorts:    []string{"6379/tcp"},
			WaitingFor:      wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, redisC)
	require.Error(t, err)
	require.Contains(t, err.Error(), "manifest unknown")
}

// setAuthConfig sets the DOCKER_AUTH_CONFIG environment variable with
// authentication for with the given host, username and password.
func setAuthConfig(t *testing.T, host, username, password string) {
	t.Helper()

	authConfigs, err := registry.DockerAuthConfig(host, username, password)
	require.NoError(t, err)
	auth, err := json.Marshal(dockercfg.Config{AuthConfigs: authConfigs})
	require.NoError(t, err)

	t.Setenv("DOCKER_AUTH_CONFIG", string(auth))
}
