package compose

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

// TestDockerComposeGoroutineLeak verifies that calling Up() followed by Down()
// and Close() does not leak net/http persistConn goroutines from the internal
// Docker CLI HTTP transport.
//
// Before the fix, two goroutines per Up+Down cycle were leaked permanently:
//
//	net/http.(*persistConn).readLoop
//	net/http.(*persistConn).writeLoop
//
// Note: Reaper (Ryuk) goroutines are intentionally excluded from this check;
// they are a separate pre-existing issue tracked in the same issue but require
// a distinct fix involving Reaper termination signal handling in Down().
func TestDockerComposeGoroutineLeak(t *testing.T) {
	path, _ := RenderComposeSimple(t)

	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	// Snapshot goroutines after NewDockerCompose establishes its initial
	// Docker provider connection, so IgnoreCurrent covers the provider's
	// keep-alive transport goroutines that pre-date the compose Up/Down cycle.
	ignoreExisting := goleak.IgnoreCurrent()

	// Register Close cleanup first so it runs last (t.Cleanup is LIFO).
	// goleak.VerifyNone is called here so it runs after both Down and Close.
	t.Cleanup(func() {
		require.NoError(t, compose.Close(), "compose.Close()")
		goleak.VerifyNone(t,
			ignoreExisting,
			// TODO(#2008): Remove this ignore when the Reaper goroutine leak is fixed.
			// This references an internal anonymous closure that may change if Reaper is refactored.
			goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go.(*Reaper).connect.func1"),
		)
	})

	ctx := context.Background()

	require.NoError(t, compose.Up(ctx, Wait(true)), "compose.Up()")

	// Register Down cleanup after Up so it runs before Close (t.Cleanup is LIFO).
	// This ensures Down is called even if the test panics after Up succeeds.
	t.Cleanup(func() {
		require.NoError(t, compose.Down(ctx, RemoveOrphans(true), RemoveVolumes(true)), "compose.Down()")
	})
}
