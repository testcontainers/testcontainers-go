package testcontainers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuguy83/dockercfg"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/wait"
)

const exampleAuth = "https://example-auth.com"

var testDockerConfigDirPath = filepath.Join("testdata", ".docker")

var indexDockerIO = core.IndexDockerIO

func TestGetDockerConfig(t *testing.T) {
	const expectedErrorMessage = "Expected to find %s in auth configs"

	t.Setenv("DOCKER_CONFIG", t.TempDir())

	// Verify that the default docker config file exists before any test in this suite runs.
	// Then, we can safely run the tests that rely on it. If it does not exist, we create it
	// using the content of the testdata/.docker/config.json file into a temporary directory.
	defaultCfg, err := dockercfg.LoadDefaultConfig()
	if err != nil {
		// create docker config file
		bs, err := os.ReadFile(filepath.Join(testDockerConfigDirPath, "config.json"))
		require.NoError(t, err)

		defaultCfgPath, err := dockercfg.ConfigPath()
		require.NoError(t, err)

		// write file to default location
		err = os.WriteFile(defaultCfgPath, bs, 0644)
		require.NoError(t, err)

		defaultCfg, err = dockercfg.LoadDefaultConfig()
		require.NoError(t, err)
	}
	require.NotEmpty(t, defaultCfg)

	t.Run("without DOCKER_CONFIG env var retrieves default", func(tt *testing.T) {
		cfg, err := getDockerConfig()
		require.NoError(tt, err)
		require.NotEmpty(tt, cfg)

		assert.Equal(tt, defaultCfg, cfg)
	})

	t.Run("with DOCKER_CONFIG env var pointing to a non-existing file raises error", func(tt *testing.T) {
		tt.Setenv("DOCKER_CONFIG", filepath.Join(testDockerConfigDirPath, "non-existing"))

		cfg, err := getDockerConfig()
		require.Error(tt, err)
		require.Empty(tt, cfg)
	})

	t.Run("with DOCKER_CONFIG env var", func(tt *testing.T) {
		tt.Setenv("DOCKER_CONFIG", testDockerConfigDirPath)

		cfg, err := getDockerConfig()
		require.NoError(tt, err)
		require.NotEmpty(tt, cfg)

		assert.Len(tt, cfg.AuthConfigs, 3)

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs[indexDockerIO]; !ok {
			tt.Errorf(expectedErrorMessage, indexDockerIO)
		}
		if _, ok := authCfgs["https://example.com"]; !ok {
			tt.Errorf(expectedErrorMessage, "https://example.com")
		}
		if _, ok := authCfgs["https://my.private.registry"]; !ok {
			tt.Errorf(expectedErrorMessage, "https://my.private.registry")
		}
	})

	t.Run("DOCKER_AUTH_CONFIG env var takes precedence", func(tt *testing.T) {
		tt.Setenv("DOCKER_AUTH_CONFIG", `{
			"auths": {
					"`+exampleAuth+`": {}
			},
			"credsStore": "desktop"
		}`)
		tt.Setenv("DOCKER_CONFIG", testDockerConfigDirPath)

		cfg, err := getDockerConfig()
		require.NoError(tt, err)
		require.NotEmpty(tt, cfg)

		assert.Len(tt, cfg.AuthConfigs, 1)

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs[indexDockerIO]; ok {
			tt.Errorf("Not expected to find %s in auth configs", indexDockerIO)
		}
		if _, ok := authCfgs[exampleAuth]; !ok {
			tt.Errorf(expectedErrorMessage, exampleAuth)
		}
	})

	t.Run("retrieve auth with DOCKER_AUTH_CONFIG env var", func(tt *testing.T) {
		base64 := "Z29waGVyOnNlY3JldA==" // gopher:secret

		tt.Setenv("DOCKER_AUTH_CONFIG", `{
			"auths": {
					"`+exampleAuth+`": { "username": "gopher", "password": "secret", "auth": "`+base64+`" }
			},
			"credsStore": "desktop"
		}`)

		registry, cfg, err := DockerImageAuth(context.Background(), exampleAuth+"/my/image:latest")
		require.NoError(tt, err)
		require.NotEmpty(tt, cfg)

		assert.Equal(tt, exampleAuth, registry)
		assert.Equal(tt, "gopher", cfg.Username)
		assert.Equal(tt, "secret", cfg.Password)
		assert.Equal(tt, base64, cfg.Auth)
	})

	t.Run("match registry authentication by host", func(tt *testing.T) {
		base64 := "Z29waGVyOnNlY3JldA==" // gopher:secret
		imageReg := "example-auth.com"
		imagePath := "/my/image:latest"

		tt.Setenv("DOCKER_AUTH_CONFIG", `{
			"auths": {
					"`+exampleAuth+`": { "username": "gopher", "password": "secret", "auth": "`+base64+`" }
			},
			"credsStore": "desktop"
		}`)

		registry, cfg, err := DockerImageAuth(context.Background(), imageReg+imagePath)
		require.NoError(tt, err)
		require.NotEmpty(tt, cfg)

		assert.Equal(tt, imageReg, registry)
		assert.Equal(tt, "gopher", cfg.Username)
		assert.Equal(tt, "secret", cfg.Password)
		assert.Equal(tt, base64, cfg.Auth)
	})

	t.Run("fail to match registry authentication due to invalid host", func(tt *testing.T) {
		base64 := "Z29waGVyOnNlY3JldA==" // gopher:secret
		imageReg := "example-auth.com"
		imagePath := "/my/image:latest"
		invalidRegistryURL := "://invalid-host"

		tt.Setenv("DOCKER_AUTH_CONFIG", `{
			"auths": {
					"`+invalidRegistryURL+`": { "username": "gopher", "password": "secret", "auth": "`+base64+`" }
			},
			"credsStore": "desktop"
		}`)

		registry, cfg, err := DockerImageAuth(context.Background(), imageReg+imagePath)
		require.Equal(tt, err, dockercfg.ErrCredentialsNotFound)
		require.Empty(tt, cfg)

		assert.Equal(tt, imageReg, registry)
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

	redisC, err := prepareRedisImage(ctx, req, t)
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, redisC)
}

// removeImageFromLocalCache removes the image from the local cache
func removeImageFromLocalCache(t *testing.T, image string) {
	ctx := context.Background()

	testcontainersClient, err := NewDockerClientWithOpts(ctx, client.WithVersion(daemonMaxVersion))
	if err != nil {
		t.Log("could not create client to cleanup registry: ", err)
	}
	defer testcontainersClient.Close()

	_, err = testcontainersClient.ImageRemove(ctx, image, types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	if err != nil {
		t.Logf("could not remove image %s: %v\n", image, err)
	}
}

func TestBuildContainerFromDockerfileWithDockerAuthConfig(t *testing.T) {
	// using the same credentials as in the Docker Registry
	base64 := "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk" // testuser:testpassword
	t.Setenv("DOCKER_AUTH_CONFIG", `{
		"auths": {
				"localhost:5001": { "username": "testuser", "password": "testpassword", "auth": "`+base64+`" }
		},
		"credsStore": "desktop"
	}`)

	prepareLocalRegistryWithAuth(t)

	ctx := context.Background()

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata",
			Dockerfile: "auth.Dockerfile",
		},
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisC, err := prepareRedisImage(ctx, req, t)
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, redisC)
}

func TestBuildContainerFromDockerfileShouldFailWithWrongDockerAuthConfig(t *testing.T) {
	// using different credentials than in the Docker Registry
	base64 := "Zm9vOmJhcg==" // foo:bar
	t.Setenv("DOCKER_AUTH_CONFIG", `{
		"auths": {
			"localhost:5001": { "username": "foo", "password": "bar", "auth": "`+base64+`" }
		},
		"credsStore": "desktop"
	}`)

	prepareLocalRegistryWithAuth(t)

	ctx := context.Background()

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata",
			Dockerfile: "auth.Dockerfile",
		},
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisC, err := prepareRedisImage(ctx, req, t)
	require.Error(t, err)
	terminateContainerOnEnd(t, ctx, redisC)
}

func TestCreateContainerFromPrivateRegistry(t *testing.T) {
	// using the same credentials as in the Docker Registry
	base64 := "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk" // testuser:testpassword
	t.Setenv("DOCKER_AUTH_CONFIG", `{
		"auths": {
				"localhost:5001": { "username": "testuser", "password": "testpassword", "auth": "`+base64+`" }
		},
		"credsStore": "desktop"
	}`)

	prepareLocalRegistryWithAuth(t)

	ctx := context.Background()
	req := ContainerRequest{
		Image:           "localhost:5001/redis:5.0-alpine",
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, redisContainer)
}

func prepareLocalRegistryWithAuth(t *testing.T) {
	ctx := context.Background()
	wd, err := os.Getwd()
	require.NoError(t, err)
	// copyDirectoryToContainer {
	req := ContainerRequest{
		Image:        "registry:2",
		ExposedPorts: []string{"5001:5000/tcp"},
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
		WaitingFor: wait.ForExposedPort(),
	}
	// }

	genContainerReq := GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	}

	registryC, err := GenericContainer(ctx, genContainerReq)
	require.NoError(t, err)

	t.Cleanup(func() {
		removeImageFromLocalCache(t, "localhost:5001/redis:5.0-alpine")
	})
	t.Cleanup(func() {
		require.NoError(t, registryC.Terminate(context.Background()))
	})

	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
}

func prepareRedisImage(ctx context.Context, req ContainerRequest, t *testing.T) (Container, error) {
	genContainerReq := GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	}

	redisC, err := GenericContainer(ctx, genContainerReq)

	return redisC, err
}
