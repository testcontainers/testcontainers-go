package testcontainers_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/testcontainers/testcontainers-go"
)

const (
	nginxAlpineImage = "docker.io/nginx:alpine"
	nginxDefaultPort = "80/tcp"
)

func terminateContainerOnEnd(tb testing.TB, ctx context.Context, ctr testcontainers.Container) {
	tb.Helper()
	if ctr == nil {
		return
	}
	tb.Cleanup(func() {
		tb.Log("terminating container")
		assert.NilError(tb, ctr.Terminate(ctx))
	})
}
