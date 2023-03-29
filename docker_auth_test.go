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
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/wait"
)

const exampleAuth = "https://example-auth.com"

var testDockerConfigDirPath = filepath.Join("testdata", ".docker")

var indexDockerIO = testcontainersdocker.IndexDockerIO

func TestGetDockerConfig(t *testing.T) {
	const expectedErrorMessage = "Expected to find %s in auth configs"

	// Verify that the default docker config file exists before any test in this suite runs.
	// Then, we can safely run the tests that rely on it.
	cfg, err := dockercfg.LoadDefaultConfig()
	require.Nil(t, err)
	require.NotNil(t, cfg)

	t.Run("without DOCKER_CONFIG env var retrieves default", func(t *testing.T) {
		cfg, err := getDockerConfig()
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 1, len(cfg.AuthConfigs))

		authCfgs := cfg.AuthConfigs

		if _, ok := authCfgs[indexDockerIO]; !ok {
			t.Errorf(expectedErrorMessage, indexDockerIO)
		}
	})

	t.Run("with DOCKER_CONFIG env var pointing to a non-existing file raises error", func(t *testing.T) {
		t.Setenv("DOCKER_CONFIG", filepath.Join(testDockerConfigDirPath, "non-existing"))

		cfg, err := getDockerConfig()
		require.NotNil(t, err)
		require.Empty(t, cfg)
	})

	t.Run("with DOCKER_CONFIG env var", func(t *testing.T) {
		t.Setenv("DOCKER_CONFIG", testDockerConfigDirPath)

		cfg, err := getDockerConfig()
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 3, len(cfg.AuthConfigs))

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
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, 1, len(cfg.AuthConfigs))

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
		require.Nil(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, exampleAuth, registry)
		assert.Equal(t, "gopher", cfg.Username)
		assert.Equal(t, "secret", cfg.Password)
		assert.Equal(t, base64, cfg.Auth)
	})
}

func TestBuildContainerFromDockerfile(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context: "./testdata",
		},
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisC, err := prepareRedisImage(ctx, req, t)
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, redisC)
}

func TestBuildContainerFromDockerfileWithDockerAuthConfig(t *testing.T) {
	// using the same credentials as in the Docker Registry
	base64 := "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk" // testuser:testpassword
	t.Setenv("DOCKER_AUTH_CONFIG", `{
		"auths": {
				"localhost:5000": { "username": "testuser", "password": "testpassword", "auth": "`+base64+`" }
		},
		"credsStore": "desktop"
	}`)

	prepareLocalRegistryWithAuth(t)
	defer func() {
		ctx := context.Background()
		testcontainersClient, err := client.NewClientWithOpts(client.WithVersion(daemonMaxVersion))
		if err != nil {
			t.Log("could not create client to cleanup registry: ", err)
		}

		_, err = testcontainersClient.ImageRemove(ctx, "localhost:5000/redis:5.0-alpine", types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if err != nil {
			t.Log("could not remove image: ", err)
		}

	}()

	ctx := context.Background()

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata",
			Dockerfile: "auth.Dockerfile",
		},

		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
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
			"localhost:5000": { "username": "foo", "password": "bar", "auth": "`+base64+`" }
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
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
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
				"localhost:5000": { "username": "testuser", "password": "testpassword", "auth": "`+base64+`" }
		},
		"credsStore": "desktop"
	}`)

	prepareLocalRegistryWithAuth(t)

	ctx := context.Background()
	req := ContainerRequest{
		Image:           "localhost:5000/redis:5.0-alpine",
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.Nil(t, err)
	terminateContainerOnEnd(t, ctx, redisContainer)
}

func prepareLocalRegistryWithAuth(t *testing.T) {
	ctx := context.Background()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	req := ContainerRequest{
		Image:        "registry:2",
		ExposedPorts: []string{"5000:5000/tcp"},
		Env: map[string]string{
			"REGISTRY_AUTH":                             "htpasswd",
			"REGISTRY_AUTH_HTPASSWD_REALM":              "Registry",
			"REGISTRY_AUTH_HTPASSWD_PATH":               "/auth/htpasswd",
			"REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY": "/data",
		},
		Mounts: ContainerMounts{
			ContainerMount{
				Source: GenericBindMountSource{
					HostPath: fmt.Sprintf("%s/testdata/auth", wd),
				},
				Target: "/auth",
			},
			ContainerMount{
				Source: GenericBindMountSource{
					HostPath: fmt.Sprintf("%s/testdata/data", wd),
				},
				Target: "/data",
			},
		},
		WaitingFor: wait.ForExposedPort(),
	}

	genContainerReq := GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	}

	registryC, err := GenericContainer(ctx, genContainerReq)
	assert.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, registryC.Terminate(context.Background()))
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
