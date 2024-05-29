package testcontainers_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
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

// TerminateContainerOnEnd is a helper function to terminate a container when the test ends.
// It will use the testing.TB.Cleanup function to terminate the container.
func TerminateContainerOnEnd(tb testing.TB, ctx context.Context, ctr testcontainers.CreatedContainer) {
	tb.Helper()
	testcontainers.TerminateContainerOnEnd(tb, ctx, ctr)
}
