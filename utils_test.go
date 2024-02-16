package testcontainers_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

const (
	nginxAlpineImage = "docker.io/nginx:alpine"
	nginxDefaultPort = "80/tcp"
)

var providerType = testcontainers.ProviderDocker

func init() {
	if strings.Contains(os.Getenv("DOCKER_HOST"), "podman.sock") {
		providerType = testcontainers.ProviderPodman
	}
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
