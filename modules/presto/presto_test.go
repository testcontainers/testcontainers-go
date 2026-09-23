package presto_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/presto"
)

func TestPresto(t *testing.T) {
	ctx := context.Background()

	ctr, err := presto.Run(ctx, "prestodb/presto:0.286")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("ConnectionString", func(t *testing.T) {
		connStr, err := ctr.ConnectionString(ctx)
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(connStr, "http://"), "expected http:// prefix, got: %s", connStr)
	})
}
