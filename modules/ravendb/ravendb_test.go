package ravendb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/ravendb"
)

func TestRavenDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := ravendb.Run(ctx, "ravendb/ravendb:6.0-ubuntu-latest")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("ManagementURL", func(t *testing.T) {
		url, err := ctr.ManagementURL(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, url)
		require.Contains(t, url, "http://")
	})

	t.Run("TCPEndpoint", func(t *testing.T) {
		endpoint, err := ctr.TCPEndpoint(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, endpoint)
	})
}
