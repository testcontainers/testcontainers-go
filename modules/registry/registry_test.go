package registry_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cpuguy83/dockercfg"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

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

	t.Run("HTTP connection without basic auth fails", func(t *testing.T) {
		httpCli := http.Client{}
		req, err := http.NewRequest(http.MethodGet, httpAddress+"/v2/_catalog", nil)
		require.NoError(t, err)

		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("HTTP connection with incorrect basic auth fails", func(t *testing.T) {
		httpCli := http.Client{}
		req, err := http.NewRequest(http.MethodGet, httpAddress+"/v2/_catalog", nil)
		require.NoError(t, err)

		req.SetBasicAuth("foo", "bar")

		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("HTTP connection with basic auth succeeds", func(t *testing.T) {
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

		redisC, err := testcontainers.Run(
			context.Background(), "",
			testcontainers.WithDockerfile(testcontainers.FromDockerfile{
				Context: filepath.Join("testdata", "redis"),
				BuildArgs: map[string]*string{
					"REGISTRY_HOST": &registryHost,
				},
			}),
			testcontainers.WithAlwaysPull(), // make sure the authentication takes place
			testcontainers.WithExposedPorts("6379/tcp"),
			testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections")),
		)
		testcontainers.CleanupContainer(tt, redisC)
		require.Error(tt, err)
		msg := strings.ToLower(err.Error())
		require.True(tt, strings.Contains(msg, "unauthorized: authentication required") || strings.Contains(msg, "401 unauthorized"), "unexpected error: %v", err)
	})

	t.Run("build image with valid credentials", func(tt *testing.T) {
		setAuthConfig(tt, registryHost, "testuser", "testpassword")

		// build a custom redis image from the private registry,
		// using RegistryName of the container as the registry.
		// The container should start because the authentication
		// is correct.

		redisC, err := testcontainers.Run(
			context.Background(), "",
			testcontainers.WithDockerfile(testcontainers.FromDockerfile{
				Context: filepath.Join("testdata", "redis"),
				BuildArgs: map[string]*string{
					"REGISTRY_HOST": &registryHost,
				},
			}),
			testcontainers.WithAlwaysPull(), // make sure the authentication takes place
			testcontainers.WithExposedPorts("6379/tcp"),
			testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections")),
		)
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

func TestRunContainer_authenticated_htpasswd_atomic_per_container(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := require.New(t)

	type container struct {
		pass     string
		registry *registry.RegistryContainer
		addr     string
	}

	newContainer := func(password string) container {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 5)
		r.NoError(err)

		reg, err := registry.Run(
			ctx,
			registry.DefaultImage,
			registry.WithHtpasswd("testuser:"+string(hash)),
		)
		r.NoError(err)
		testcontainers.CleanupContainer(t, reg)

		addr, err := reg.Address(ctx)
		r.NoError(err)

		return container{pass: password, registry: reg, addr: addr}
	}

	// Create two independent registries with different credentials.
	regA := newContainer("passA")
	regB := newContainer("passB")

	client := http.Client{}

	// 1. Wrong password against A must fail.
	req, err := http.NewRequest(http.MethodGet, regA.addr+"/v2/_catalog", nil)
	r.NoError(err)
	req.SetBasicAuth("testuser", regB.pass)
	resp, err := client.Do(req)
	r.NoError(err)
	r.Equal(http.StatusUnauthorized, resp.StatusCode)
	_ = resp.Body.Close()

	// 2. Correct password against A must succeed.
	req.SetBasicAuth("testuser", regA.pass)
	resp, err = client.Do(req)
	r.NoError(err)
	r.Equal(http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// 3. Correct password against B must succeed.
	req, err = http.NewRequest(http.MethodGet, regB.addr+"/v2/_catalog", nil)
	r.NoError(err)
	req.SetBasicAuth("testuser", regB.pass)
	resp, err = client.Do(req)
	r.NoError(err)
	r.Equal(http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
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

	redisC, err := testcontainers.Run(context.Background(), "",
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context: filepath.Join("testdata", "redis"),
			BuildArgs: map[string]*string{
				"REGISTRY_HOST": &registryHost,
			},
		}),
		testcontainers.WithAlwaysPull(), // make sure the authentication takes place
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections")),
	)
	testcontainers.CleanupContainer(t, redisC)
	require.Error(t, err)

	msg := strings.ToLower(err.Error())
	require.True(t, strings.Contains(msg, "manifest unknown") || strings.Contains(msg, "no such image") || strings.Contains(msg, registryHost+"/redis:5.0-alpine: not found"), "unexpected error: %v", err)
}

func TestPullImage_samePlatform(t *testing.T) {
	ctx := context.Background()
	registryContainer, err := registry.Run(ctx, registry.DefaultImage)
	testcontainers.CleanupContainer(t, registryContainer)
	require.NoError(t, err)

	inspect, err := registryContainer.Inspect(ctx)
	require.NoError(t, err)

	dockerCli, err := testcontainers.NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer dockerCli.Close()

	// Pull an image that shares the same platform as the registry container's image.
	const img = "redis:latest"

	err = registryContainer.PullImage(ctx, img)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := dockerCli.ImageRemove(ctx, img, client.ImageRemoveOptions{Force: true})
		require.NoError(t, err)
	})

	imgInspect, err := dockerCli.ImageInspect(ctx, img)
	require.NoError(t, err)

	if inspect.ImageManifestDescriptor != nil && inspect.ImageManifestDescriptor.Platform != nil {
		require.Equal(t, inspect.ImageManifestDescriptor.Platform.Architecture, imgInspect.Architecture)
		require.Equal(t, inspect.ImageManifestDescriptor.Platform.OS, imgInspect.Os)
	} else {
		require.Equal(t, "linux", imgInspect.Os)
		require.Equal(t, runtime.GOARCH, imgInspect.Architecture)
	}
}

func TestRegistry_PushImage(t *testing.T) {
	ctx := context.Background()
	ctr, err := registry.Run(ctx, registry.DefaultImage,
		registry.WithHtpasswdFile(filepath.Join("testdata", "auth", "htpasswd")),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	host, err := ctr.HostAddress(ctx)
	require.NoError(t, err)
	setAuthConfig(t, host, "testuser", "testpassword")

	// Build locally so containerd produces an OCI manifest instead of relying
	// on the Docker schema 2 manifests in the pre-populated registry fixture.
	buildContext := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(buildContext, "Dockerfile"),
		[]byte("FROM scratch\nLABEL test=registry-push\nCMD [\"/unused\"]\n"), 0o600))
	repo := host + "/push-test"
	imageRef := repo + ":test"
	imageContainer, err := testcontainers.Run(ctx, "",
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context: buildContext,
			Repo:    repo,
			Tag:     "test",
		}),
		testcontainers.WithNoStart(),
	)
	testcontainers.CleanupContainer(t, imageContainer)
	require.NoError(t, err)

	t.Run("push and find image", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		require.NoError(t, ctr.PushImage(ctx, imageRef))
		require.NoError(t, ctr.ImageExists(ctx, imageRef))
	})

	t.Run("push with wrong credentials reports the push error", func(t *testing.T) {
		setAuthConfig(t, host, "testuser", "wrong-password")
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		err := ctr.PushImage(ctx, imageRef)
		require.ErrorContains(t, err, "push image")
		require.Contains(t, strings.ToLower(err.Error()), "unauthorized")
		require.NotErrorIs(t, err, context.DeadlineExceeded, "push errors must not be hidden by a readiness timeout")
	})

	t.Run("delete image", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		require.NoError(t, ctr.DeleteImage(ctx, imageRef))

		address, err := ctr.Address(ctx)
		require.NoError(t, err)
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, address+"/v2/push-test/manifests/test", nil)
		require.NoError(t, err)
		req.SetBasicAuth("testuser", "testpassword")
		req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
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
