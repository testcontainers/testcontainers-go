package couchdb_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/couchdb"
)

func TestCouchDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := couchdb.Run(ctx, "couchdb:3")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("connection string", func(t *testing.T) {
		connStr, err := ctr.ConnectionString(ctx)
		require.NoError(t, err)
		require.Contains(t, connStr, "http://admin:password@")
		require.Contains(t, connStr, ":5984")
	})

	t.Run("http endpoint reachable", func(t *testing.T) {
		connStr, err := ctr.ConnectionString(ctx)
		require.NoError(t, err)

		resp, err := http.Get(connStr + "/_up")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestCouchDB_WithAdminCredentials(t *testing.T) {
	ctx := context.Background()

	ctr, err := couchdb.Run(ctx, "couchdb:3",
		couchdb.WithAdminCredentials("testuser", "testpassword"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.Contains(t, connStr, "http://testuser:testpassword@")
}
