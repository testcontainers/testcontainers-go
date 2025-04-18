package testcontainers

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	dockerconfig "github.com/testcontainers/testcontainers-go/internal/docker/config"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	exampleAuth = "https://example-auth.com"
)

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
		require.ErrorIs(t, err, dockerconfig.ErrCredentialsNotFound)
		require.Empty(t, cfg)
		require.Equal(t, imageReg, registry)
	})

	t.Run("fail to match registry authentication by host with empty URL scheme creds and missing default", func(t *testing.T) {
		origDefaultRegistryFn := defaultRegistryFn
		t.Cleanup(func() {
			defaultRegistryFn = origDefaultRegistryFn
		})
		defaultRegistryFn = func(_ context.Context) string {
			return ""
		}

		imageReg := ""
		imagePath := "image:latest"

		setAuthConfig(t, "example-auth.com", "gopher", "secret")

		registry, cfg, err := DockerImageAuth(context.Background(), imageReg+imagePath)
		require.ErrorIs(t, err, dockerconfig.ErrCredentialsNotFound)
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
	t.Helper()
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
	t.Helper()
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
				HostFilePath:      wd + "/testdata/auth",
				ContainerFilePath: "/auth",
			},
			{
				HostFilePath:      wd + "/testdata/data",
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
	t.Helper()
	if os.Getenv("WSL_DISTRO_NAME") == "" {
		return "localhost"
	}

	conn, err := net.Dial("udp", "golang.org:80")
	require.NoError(t, err)
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
