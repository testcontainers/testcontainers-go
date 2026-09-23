package questdb_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/questdb"
)

func TestQuestDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := questdb.Run(ctx, "questdb/questdb:7.4.2")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("HTTPEndpoint", func(t *testing.T) {
		endpoint, err := ctr.HTTPEndpoint(ctx)
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(endpoint, "http://"), "expected http:// prefix, got: %s", endpoint)
	})

	t.Run("PGEndpoint", func(t *testing.T) {
		endpoint, err := ctr.PGEndpoint(ctx)
		require.NoError(t, err)
		// admin:quest are the built-in QuestDB default credentials
		require.True(t, strings.HasPrefix(endpoint, "postgres://admin:quest@"), "expected postgres://admin:quest@ prefix, got: %s", endpoint)
		require.True(t, strings.HasSuffix(endpoint, "/qdb"), "expected /qdb suffix, got: %s", endpoint)
	})

	t.Run("InfluxDBEndpoint", func(t *testing.T) {
		endpoint, err := ctr.InfluxDBEndpoint(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, endpoint)
		require.Contains(t, endpoint, ":")
	})
}
