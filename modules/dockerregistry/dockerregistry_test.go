package dockerregistry

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var originalDockerAuthConfig string

func init() {
	originalDockerAuthConfig = os.Getenv("DOCKER_AUTH_CONFIG")
}

func TestDockerRegistry(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	port, ipAddress := getRegistryPortAndAddress(t, err, container, ctx)

	// Let's simply check that the registry is up and running with a GET to http://localhost:5000/v2/_catalog
	resp, err := http.Get("http://" + ipAddress + ":" + port.Port() + "/v2/_catalog")
	if err != nil {
		// handle err
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
}

func TestDockerRegistryWithCustomImage(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, WithImage("docker.io/registry:latest"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	port, ipAddress := getRegistryPortAndAddress(t, err, container, ctx)

	// Let's check that the registry is up and running with a GET to http://localhost:5000/v2/_catalog also using a different image
	resp, err := http.Get("http://" + ipAddress + ":" + port.Port() + "/v2/_catalog")
	if err != nil {
		// handle err
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
}

func TestDockerRegistryWithData(t *testing.T) {
	ctx := context.Background()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	container, err := RunContainer(ctx, WithData(wd+"/../../testdata/data"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// Let's check that we are able to start a container form image localhost:5000/redis:5.0-alpine
	req := testcontainers.ContainerRequest{
		Image:           "localhost:5000/redis:5.0-alpine",
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.Nil(t, err)
	terminateContainerOnEnd(t, ctx, redisContainer)
	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
}

func TestDockerRegistryWithAuth(t *testing.T) {
	ctx := context.Background()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	container, err := RunContainer(ctx, WithAuthentication(wd+"/../../testdata/auth"))

	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	port, ipAddress := getRegistryPortAndAddress(t, err, container, ctx)

	// Let's simply check that the registry is up and running with a GET to http://localhost:5000/v2/_catalog
	h := http.Client{}
	req, _ := http.NewRequest("GET", "http://"+ipAddress+":"+port.Port()+"/v2/_catalog", nil)
	req.SetBasicAuth("testuser", "testpassword")
	resp, _ := h.Do(req)
	defer resp.Body.Close()
	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
}

func TestDockerRegistryWithAuthWithUnauthorizedRequest(t *testing.T) {
	ctx := context.Background()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	container, err := RunContainer(ctx, WithAuthentication(wd+"/../../testdata/auth"))

	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	port, ipAddress := getRegistryPortAndAddress(t, err, container, ctx)

	// Let's simply check that the registry is up and running with a GET to http://localhost:5000/v2/_catalog
	h := http.Client{}
	req, _ := http.NewRequest("GET", "http://"+ipAddress+":"+port.Port()+"/v2/_catalog", nil)
	resp, err := h.Do(req)
	require.Equal(t, resp.StatusCode, 401)
	defer resp.Body.Close()
	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
}

func TestDockerRegistryWithAuthAndData(t *testing.T) {
	t.Cleanup(func() {
		os.Setenv("DOCKER_AUTH_CONFIG", originalDockerAuthConfig)
	})
	os.Unsetenv("DOCKER_AUTH_CONFIG")

	// using the same credentials as in the Docker Registry
	base64 := "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk" // testuser:testpassword
	t.Setenv("DOCKER_AUTH_CONFIG", `{
		"auths": {
				"localhost:5000": { "username": "testuser", "password": "testpassword", "auth": "`+base64+`" }
		},
		"credsStore": "desktop"
	}`)
	ctx := context.Background()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	container, err := RunContainer(ctx, WithAuthentication(wd+"/../../testdata/auth"), WithData(wd+"/../../testdata/data"))

	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// Let's check that we are able to start a container form image localhost:5000/redis:5.0-alpine
	// using default username and password
	req := testcontainers.ContainerRequest{
		Image:           "localhost:5000/redis:5.0-alpine",
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.Nil(t, err)
	terminateContainerOnEnd(t, ctx, redisContainer)
	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
}

func TestDockerRegistryWithAuthDataAndImage(t *testing.T) {
	t.Cleanup(func() {
		os.Setenv("DOCKER_AUTH_CONFIG", originalDockerAuthConfig)
	})
	os.Unsetenv("DOCKER_AUTH_CONFIG")

	// using the same credentials as in the Docker Registry
	base64 := "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk" // testuser:testpassword
	t.Setenv("DOCKER_AUTH_CONFIG", `{
		"auths": {
				"localhost:5000": { "username": "testuser", "password": "testpassword", "auth": "`+base64+`" }
		},
		"credsStore": "desktop"
	}`)
	ctx := context.Background()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	container, err := RunContainer(ctx, WithAuthentication(wd+"/../../testdata/auth"), WithData(wd+"/../../testdata/data"), WithImage("docker.io/registry:latest"))

	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// Let's check that we are able to start a container form image localhost:5000/redis:5.0-alpine
	// using default username and password
	req := testcontainers.ContainerRequest{
		Image:           "localhost:5000/redis:5.0-alpine",
		AlwaysPullImage: true, // make sure the authentication takes place
		ExposedPorts:    []string{"6379/tcp"},
		WaitingFor:      wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.Nil(t, err)
	terminateContainerOnEnd(t, ctx, redisContainer)
	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
}
func getRegistryPortAndAddress(t *testing.T, err error, container *DockerRegistryContainer, ctx context.Context) (nat.Port, string) {
	port, err := container.MappedPort(ctx, "5000")
	if err != nil {
		t.Fatal(err)
	}

	ipAddress, err := container.Host(ctx)

	if err != nil {
		t.Fatal(err)
	}
	return port, ipAddress
}

func terminateContainerOnEnd(tb testing.TB, ctx context.Context, ctr testcontainers.Container) {
	tb.Helper()
	if ctr == nil {
		return
	}
	tb.Cleanup(func() {
		tb.Log("terminating container")
		require.NoError(tb, ctr.Terminate(ctx))
	})
}
