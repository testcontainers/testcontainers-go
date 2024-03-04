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
	mysqlImage        = "docker.io/mysql:8.0.36"
	nginxDelayedImage = "docker.io/menedev/delayed-nginx:1.15.2"
	nginxImage        = "docker.io/nginx"
	nginxAlpineImage  = "docker.io/nginx:alpine"
	nginxDefaultPort  = "80/tcp"
	nginxHighPort     = "8080/tcp"
	daemonMaxVersion  = "1.41"
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
