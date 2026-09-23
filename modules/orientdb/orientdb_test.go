package orientdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/orientdb"
)

func TestOrientDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := orientdb.Run(ctx, "orientdb:3.2")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("ServerURL", func(t *testing.T) {
		serverURL, err := ctr.ServerURL(ctx)
		require.NoError(t, err)
		require.Contains(t, serverURL, "remote:")
	})

	t.Run("StudioURL", func(t *testing.T) {
		studioURL, err := ctr.StudioURL(ctx)
		require.NoError(t, err)
		require.Contains(t, studioURL, "http://")
	})
}

func TestOrientDBWithRootPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := orientdb.Run(ctx, "orientdb:3.2",
		orientdb.WithRootPassword("mysecret"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	studioURL, err := ctr.StudioURL(ctx)
	require.NoError(t, err)
	require.Contains(t, studioURL, "http://")
}
