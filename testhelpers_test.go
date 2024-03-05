package testcontainers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

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
