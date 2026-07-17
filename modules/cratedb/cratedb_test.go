package cratedb_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cratedb"
)

func TestCrateDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := cratedb.Run(ctx, "crate:5.7")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("HTTPEndpoint", func(t *testing.T) {
		endpoint, err := ctr.HTTPEndpoint(ctx)
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(endpoint, "http://"), "endpoint must start with http://: %s", endpoint)

		// Verify the Admin UI is reachable
		resp, err := http.Get(endpoint)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("PGConnectionString", func(t *testing.T) {
		connStr, err := ctr.PGConnectionString(ctx)
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(connStr, "postgres://crate@"), "connection string must start with postgres://crate@: %s", connStr)
	})

	t.Run("PGConnectionString_WithArgs", func(t *testing.T) {
		connStr, err := ctr.PGConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err)
		require.Contains(t, connStr, "sslmode=disable")
	})
}

func TestCrateDB_WithHeapSize(t *testing.T) {
	ctx := context.Background()

	ctr, err := cratedb.Run(ctx, "crate:5.7",
		cratedb.WithHeapSize("256m"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Verify the heap size env var was applied to the running container
	inspect, err := ctr.Inspect(ctx)
	require.NoError(t, err)
	require.Contains(t, inspect.Config.Env, "CRATE_HEAP_SIZE=256m", "expected CRATE_HEAP_SIZE=256m to be set in container environment")
}
