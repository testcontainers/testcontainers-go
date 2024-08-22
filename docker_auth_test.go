package testcontainers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestBuildContainerFromDockerfile(t *testing.T) {
	ctx := context.Background()
	req := Request{
		FromDockerfile: FromDockerfile{
			Context: "./testdata",
		},
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
		Started:         true,
	}

	redisC, err := Run(ctx, req)
	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, redisC)
}

// removeImageFromLocalCache removes the image from the local cache
func removeImageFromLocalCache(t *testing.T, img string) {
	ctx := context.Background()

	testcontainersClient, err := core.NewClient(ctx, client.WithVersion(daemonMaxVersion))
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

	req := Request{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata",
			Dockerfile: "auth.Dockerfile",
			BuildArgs: map[string]*string{
				"REGISTRY_HOST": &registryHost,
			},
			Repo:          "localhost",
			PrintBuildLog: true,
		},
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
		Started:         true,
	}

	redisC, err := Run(ctx, req)
	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, redisC)
}

func TestBuildContainerFromDockerfileShouldFailWithWrongDockerAuthConfig(t *testing.T) {
	registryHost := prepareLocalRegistryWithAuth(t)

	// using different credentials than in the Docker Registry
	setAuthConfig(t, registryHost, "foo", "bar")

	ctx := context.Background()

	req := Request{
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
		Started:         true,
	}

	redisC, err := Run(ctx, req)
	require.Error(t, err)
	TerminateContainerOnEnd(t, ctx, redisC)
}

func TestCreateContainerFromPrivateRegistry(t *testing.T) {
	registryHost := prepareLocalRegistryWithAuth(t)

	// using the same credentials as in the Docker Registry
	setAuthConfig(t, registryHost, "testuser", "testpassword")

	ctx := context.Background()
	req := Request{
		Image:           registryHost + "/redis:5.0-alpine",
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
		Started:         true,
	}

	redisContainer, err := Run(ctx, req)
	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, redisContainer)
}

func prepareLocalRegistryWithAuth(t *testing.T) string {
	ctx := context.Background()
	wd, err := os.Getwd()
	require.NoError(t, err)

	req := Request{
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
		WaitingFor: wait.ForExposedPort(),
		Started:    true,
	}

	registryC, err := Run(ctx, req)
	require.NoError(t, err)

	mappedPort, err := registryC.MappedPort(ctx, "5000/tcp")
	require.NoError(t, err)

	ip := localAddress(t)
	mp := mappedPort.Port()
	addr := ip + ":" + mp

	t.Cleanup(func() {
		removeImageFromLocalCache(t, addr+"/redis:5.0-alpine")
	})
	t.Cleanup(func() {
		require.NoError(t, registryC.Terminate(context.Background()))
	})

	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	return addr
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
