package testcontainers

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuguy83/dockercfg"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/wait"
)

const exampleAuth = "https://example-auth.com"

func Test_getDockerConfig(t *testing.T) {
	expectedConfig := &dockercfg.Config{
		AuthConfigs: map[string]dockercfg.AuthConfig{
			core.IndexDockerIO:            {},
			"https://example.com":         {},
			"https://my.private.registry": {},
		},
		CredentialsStore: "desktop",
	}
	t.Run("HOME/valid", func(t *testing.T) {
		testDockerConfigHome(t, "testdata")

		cfg, err := getDockerConfig()
		require.NoError(t, err)
		require.Equal(t, expectedConfig, cfg)
	})

	t.Run("HOME/not-found", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")

		cfg, err := getDockerConfig()
		require.ErrorIs(t, err, os.ErrNotExist)
		require.Nil(t, cfg)
	})

	t.Run("HOME/invalid-config", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "invalid-config")

		cfg, err := getDockerConfig()
		require.ErrorContains(t, err, "json: cannot unmarshal array")
		require.Nil(t, cfg)
	})

	t.Run("DOCKER_AUTH_CONFIG/valid", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")
		t.Setenv("DOCKER_AUTH_CONFIG", dockerConfig)

		cfg, err := getDockerConfig()
		require.NoError(t, err)
		require.Equal(t, expectedConfig, cfg)
	})

	t.Run("DOCKER_AUTH_CONFIG/invalid-config", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")
		t.Setenv("DOCKER_AUTH_CONFIG", `{"auths": []}`)

		cfg, err := getDockerConfig()
		require.ErrorContains(t, err, "json: cannot unmarshal array")
		require.Nil(t, cfg)
	})

	t.Run("DOCKER_CONFIG/valid", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")
		t.Setenv("DOCKER_CONFIG", filepath.Join("testdata", ".docker"))

		cfg, err := getDockerConfig()
		require.NoError(t, err)
		require.Equal(t, expectedConfig, cfg)
	})

	t.Run("DOCKER_CONFIG/invalid-config", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")
		t.Setenv("DOCKER_CONFIG", filepath.Join("testdata", "invalid-config", ".docker"))

		cfg, err := getDockerConfig()
		require.ErrorContains(t, err, "json: cannot unmarshal array")
		require.Nil(t, cfg)
	})
}

func TestDockerImageAuth(t *testing.T) {
	t.Run("retrieve auth with DOCKER_AUTH_CONFIG env var", func(t *testing.T) {
		username, password := "gopher", "secret"
		creds := setAuthConfig(t, exampleAuth, username, password)

		registry, cfg, err := DockerImageAuth(context.Background(), exampleAuth+"/my/image:latest")
		require.NoError(t, err)
		require.Equal(t, exampleAuth, registry)
		require.Equal(t, username, cfg.Username)
		require.Equal(t, password, cfg.Password)
		require.Equal(t, creds, cfg.Auth)
	})

	t.Run("match registry authentication by host", func(t *testing.T) {
		imageReg := "example-auth.com"
		imagePath := "/my/image:latest"
		base64 := setAuthConfig(t, exampleAuth, "gopher", "secret")

		registry, cfg, err := DockerImageAuth(context.Background(), imageReg+imagePath)
		require.NoError(t, err)
		require.Equal(t, imageReg, registry)
		require.Equal(t, "gopher", cfg.Username)
		require.Equal(t, "secret", cfg.Password)
		require.Equal(t, base64, cfg.Auth)
	})

	t.Run("fail to match registry authentication due to invalid host", func(t *testing.T) {
		imageReg := "example-auth.com"
		imagePath := "/my/image:latest"
		invalidRegistryURL := "://invalid-host"

		setAuthConfig(t, invalidRegistryURL, "gopher", "secret")

		registry, cfg, err := DockerImageAuth(context.Background(), imageReg+imagePath)
		require.ErrorIs(t, err, dockercfg.ErrCredentialsNotFound)
		require.Empty(t, cfg)
		require.Equal(t, imageReg, registry)
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

		registry, cfg, err := DockerImageAuth(context.Background(), imageReg+imagePath)
		require.ErrorIs(t, err, dockercfg.ErrCredentialsNotFound)
		require.Empty(t, cfg)
		require.Equal(t, imageReg, registry)
	})
}

func TestBuildContainerFromDockerfile(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context: "./testdata",
		},
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisC, err := prepareRedisImage(ctx, req)
	CleanupContainer(t, redisC)
	require.NoError(t, err)
}

// removeImageFromLocalCache removes the image from the local cache
func removeImageFromLocalCache(t *testing.T, img string) {
	ctx := context.Background()

	testcontainersClient, err := NewDockerClientWithOpts(ctx, client.WithVersion(daemonMaxVersion))
	if err != nil {
		t.Log("could not create client to cleanup registry: ", err)
	}
	defer testcontainersClient.Close()

	_, err = testcontainersClient.ImageRemove(ctx, img, image.RemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	if err != nil && !client.IsErrNotFound(err) {
		t.Logf("could not remove image %s: %v\n", img, err)
	}
}

func TestBuildContainerFromDockerfileWithDockerAuthConfig(t *testing.T) {
	registryHost := prepareLocalRegistryWithAuth(t)

	// using the same credentials as in the Docker Registry
	setAuthConfig(t, registryHost, "testuser", "testpassword")

	ctx := context.Background()

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata",
			Dockerfile: "auth.Dockerfile",
			BuildArgs: map[string]*string{
				"REGISTRY_HOST": &registryHost,
			},
			Repo: "localhost",
		},
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisC, err := prepareRedisImage(ctx, req)
	CleanupContainer(t, redisC)
	require.NoError(t, err)
}

func TestBuildContainerFromDockerfileShouldFailWithWrongDockerAuthConfig(t *testing.T) {
	registryHost := prepareLocalRegistryWithAuth(t)

	// using different credentials than in the Docker Registry
	setAuthConfig(t, registryHost, "foo", "bar")

	ctx := context.Background()

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata",
			Dockerfile: "auth.Dockerfile",
			BuildArgs: map[string]*string{
				"REGISTRY_HOST": &registryHost,
			},
		},
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisC, err := prepareRedisImage(ctx, req)
	CleanupContainer(t, redisC)
	require.Error(t, err)
}

func TestCreateContainerFromPrivateRegistry(t *testing.T) {
	registryHost := prepareLocalRegistryWithAuth(t)

	// using the same credentials as in the Docker Registry
	setAuthConfig(t, registryHost, "testuser", "testpassword")

	ctx := context.Background()
	req := ContainerRequest{
		Image:           registryHost + "/redis:5.0-alpine",
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	CleanupContainer(t, redisContainer)
	require.NoError(t, err)
}

func prepareLocalRegistryWithAuth(t *testing.T) string {
	ctx := context.Background()
	wd, err := os.Getwd()
	require.NoError(t, err)
	// copyDirectoryToContainer {
	req := ContainerRequest{
		Image:        "registry:2",
		ExposedPorts: []string{"5000/tcp"},
		Env: map[string]string{
			"REGISTRY_AUTH":                             "htpasswd",
			"REGISTRY_AUTH_HTPASSWD_REALM":              "Registry",
			"REGISTRY_AUTH_HTPASSWD_PATH":               "/auth/htpasswd",
			"REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY": "/data",
		},
		Files: []ContainerFile{
			{
				HostFilePath:      fmt.Sprintf("%s/testdata/auth", wd),
				ContainerFilePath: "/auth",
			},
			{
				HostFilePath:      fmt.Sprintf("%s/testdata/data", wd),
				ContainerFilePath: "/data",
			},
		},
		WaitingFor: wait.ForHTTP("/").WithPort("5000/tcp"),
	}
	// }

	genContainerReq := GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	}

	registryC, err := GenericContainer(ctx, genContainerReq)
	CleanupContainer(t, registryC)
	require.NoError(t, err)

	mappedPort, err := registryC.MappedPort(ctx, "5000/tcp")
	require.NoError(t, err)

	ip := localAddress(t)
	mp := mappedPort.Port()
	addr := ip + ":" + mp

	t.Cleanup(func() {
		removeImageFromLocalCache(t, addr+"/redis:5.0-alpine")
	})

	return addr
}

func prepareRedisImage(ctx context.Context, req ContainerRequest) (Container, error) {
	genContainerReq := GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	}

	return GenericContainer(ctx, genContainerReq)
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

// localAddress returns the local address of the machine
// which can be used to connect to the local registry.
// This avoids the issues with localhost on WSL.
func localAddress(t *testing.T) string {
	if os.Getenv("WSL_DISTRO_NAME") == "" {
		return "localhost"
	}

	conn, err := net.Dial("udp", "golang.org:80")
	require.NoError(t, err)
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

//go:embed testdata/.docker/config.json
var dockerConfig string

func Test_getDockerAuthConfigs(t *testing.T) {
	t.Run("HOME/valid", func(t *testing.T) {
		testDockerConfigHome(t, "testdata")

		requireValidAuthConfig(t)
	})

	t.Run("HOME/not-found", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-exist")

		authConfigs, err := getDockerAuthConfigs()
		require.NoError(t, err)
		require.NotNil(t, authConfigs)
		require.Empty(t, authConfigs)
	})

	t.Run("HOME/invalid-config", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "invalid-config")

		authConfigs, err := getDockerAuthConfigs()
		require.ErrorContains(t, err, "json: cannot unmarshal array")
		require.Nil(t, authConfigs)
	})

	t.Run("DOCKER_AUTH_CONFIG/valid", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-exist")
		t.Setenv("DOCKER_AUTH_CONFIG", dockerConfig)

		requireValidAuthConfig(t)
	})

	t.Run("DOCKER_AUTH_CONFIG/invalid-config", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-exist")
		t.Setenv("DOCKER_AUTH_CONFIG", `{"auths": []}`)

		authConfigs, err := getDockerAuthConfigs()
		require.ErrorContains(t, err, "json: cannot unmarshal array")
		require.Nil(t, authConfigs)
	})

	t.Run("DOCKER_CONFIG/valid", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")
		t.Setenv("DOCKER_CONFIG", filepath.Join("testdata", ".docker"))

		requireValidAuthConfig(t)
	})

	t.Run("DOCKER_CONFIG/invalid-config", func(t *testing.T) {
		testDockerConfigHome(t, "testdata", "not-found")
		t.Setenv("DOCKER_CONFIG", filepath.Join("testdata", "invalid-config", ".docker"))

		cfg, err := getDockerConfig()
		require.ErrorContains(t, err, "json: cannot unmarshal array")
		require.Nil(t, cfg)
	})
}

// requireValidAuthConfig checks that the given authConfigs map contains the expected keys.
func requireValidAuthConfig(t *testing.T) {
	t.Helper()

	authConfigs, err := getDockerAuthConfigs()
	require.NoError(t, err)

	// We can only check the keys as the values are not deterministic as they depend
	// on users environment.
	expected := map[string]registry.AuthConfig{
		"https://index.docker.io/v1/": {},
		"https://example.com":         {},
		"https://my.private.registry": {},
	}
	for k := range authConfigs {
		authConfigs[k] = registry.AuthConfig{}
	}
	require.Equal(t, expected, authConfigs)
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
