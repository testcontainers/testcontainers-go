package dockerregistry

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

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

	port, err := container.MappedPort(ctx, "5000")
	if err != nil {
		t.Fatal(err)
	}

	ipAddress, err := container.Host(ctx)

	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)
	// perform assertions
	resp, err := http.Get("http://" + ipAddress + ":" + port.Port() + "/v2/_catalog")
	if err != nil {
		// handle err
		t.Fatal(err)
	}
	defer resp.Body.Close()
}

func TestDockerRegistryWithAuth(t *testing.T) {
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

	time.Sleep(1 * time.Second)

	ctx = context.Background()
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
