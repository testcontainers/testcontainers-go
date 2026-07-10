package firebird_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/firebird"
)

func TestFirebird(t *testing.T) {
	ctx := context.Background()

	ctr, err := firebird.Run(ctx, "ghcr.io/jacobalberty/firebird:v3.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	connStr, err := ctr.ConnectionString(ctx)
	// }
	require.NoError(t, err)
	require.NotEmpty(t, connStr)
	require.Contains(t, connStr, "firebird://")
}

func TestFirebirdWithOptions(t *testing.T) {
	ctx := context.Background()

	ctr, err := firebird.Run(ctx,
		"ghcr.io/jacobalberty/firebird:v3.0",
		firebird.WithDatabase("mydb.fdb"),
		firebird.WithUsername("myuser"),
		firebird.WithPassword("mypassword"),
		firebird.WithSYSDBAPassword("mysyspassword"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.Contains(t, connStr, "myuser")
	require.Contains(t, connStr, "mypassword")
	require.Contains(t, connStr, "mydb.fdb")
}
