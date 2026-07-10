package kurrentdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kurrentdb"
)

func TestKurrentDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := kurrentdb.Run(ctx, "kurrentplatform/kurrentdb:latest")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	connStr, err := ctr.ConnectionString(ctx)
	// }
	require.NoError(t, err)
	require.NotEmpty(t, connStr)

	// Default mode is insecure, so TLS is disabled.
	require.Contains(t, connStr, "kurrent://")
	require.Contains(t, connStr, "?tls=false")
}

func TestKurrentDBWithInsecure(t *testing.T) {
	ctx := context.Background()

	ctr, err := kurrentdb.Run(ctx, "kurrentplatform/kurrentdb:latest",
		kurrentdb.WithInsecure(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.Contains(t, connStr, "?tls=false")
}
