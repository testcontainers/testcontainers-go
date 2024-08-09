package testcontainers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

const (
	mysqlImage        = "docker.io/mysql:8.0.36"
	nginxAlpineImage  = "docker.io/nginx:alpine"
	nginxDefaultPort  = "80/tcp"
	nginxDelayedImage = "docker.io/menedev/delayed-nginx:1.15.2"
	nginxImage        = "docker.io/nginx"
	nginxHighPort     = "8080/tcp"
	daemonMaxVersion  = "1.41"
)

// SkipIfContainerRuntimeIsNotHealthy is a utility function capable of skipping tests
// if the provider is not healthy, or running at all.
// This is a function designed to be used in your test, when Docker is not mandatory for CI/CD.
// In this way tests that depend on Testcontainers won't run if the provider is provisioned correctly.
func SkipIfContainerRuntimeIsNotHealthy(t *testing.T) {
	ctx := context.Background()
	cli, err := core.NewClient(ctx)
	if err != nil {
		t.Skipf("Docker is not running. TestContainers can't perform is work without it: %s", err)
	}
	defer cli.Close()

	_, err = cli.Info(ctx)
	if err != nil {
		t.Skipf("Docker is not running. TestContainers can't perform is work without it: %s", err)
	}
}

// SkipIfDockerDesktop is a utility function capable of skipping tests
// if tests are run using Docker Desktop.
func SkipIfDockerDesktop(t *testing.T, ctx context.Context) {
	cli, err := core.NewClient(ctx)
	if err != nil {
		t.Fatalf("failed to create docker client: %s", err)
	}

	info, err := cli.Info(ctx)
	if err != nil {
		t.Fatalf("failed to get docker info: %s", err)
	}

	if info.OperatingSystem == "Docker Desktop" {
		t.Skip("Skipping test that requires host network access when running in Docker Desktop")
	}
}

// TerminateContainerOnEnd is a helper function to terminate a container when the test ends.
// It will use the testing.TB.Cleanup function to terminate the container.
func TerminateContainerOnEnd(tb testing.TB, ctx context.Context, ctr CreatedContainer) {
	tb.Helper()
	if ctr == nil {
		return
	}

	// check if it's a nil DockerContainer struct
	if ctr.(*DockerContainer) == nil {
		return
	}

	tb.Cleanup(func() {
		if ctr != nil {
			require.NoError(tb, ctr.Terminate(ctx))
		}
	})
}
