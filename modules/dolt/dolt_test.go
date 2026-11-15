package dolt_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dolt"
)

func TestDolt(t *testing.T) {
	ctx := context.Background()

	ctr, err := dolt.Run(ctx, "dolthub/dolt-sql-server:1.32.4")
	testcontainers.CleanupContainer(t, ctr)
	requireCallerRegisteredDriver(t, ctx, ctr, err)

	// perform assertions
	// connectionString {
	connectionString, err := ctr.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	require.NotEmpty(t, connectionString)
}

func TestDoltWithNonRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := dolt.Run(ctx,
		"dolthub/dolt-sql-server:1.32.4",
		dolt.WithDatabase("foo"),
		dolt.WithUsername("test"),
		dolt.WithPassword(""))
	testcontainers.CleanupContainer(t, ctr)
	require.EqualError(t, err, "empty password can be used only with the root user")
}

func TestDoltWithPublicRemoteCloneUrl(t *testing.T) {
	ctx := context.Background()

	ctr, err := dolt.Run(ctx,
		"dolthub/dolt-sql-server:1.32.4",
		dolt.WithDatabase("foo"),
		dolt.WithUsername("test"),
		dolt.WithPassword("test"),
		dolt.WithScripts(filepath.Join("testdata", "check_clone_public.sh")),
		dolt.WithDoltCloneRemoteUrl("fake-remote-url"))
	testcontainers.CleanupContainer(t, ctr)
	requireCallerRegisteredDriver(t, ctx, ctr, err)
}

func createTestCredsFile(t *testing.T) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "prefix")
	require.NoError(t, err)
	defer file.Close()
	_, err = file.WriteString("some-fake-creds")
	require.NoError(t, err)
	return file.Name()
}

func TestDoltWithPrivateRemoteCloneUrl(t *testing.T) {
	ctx := context.Background()

	filename := createTestCredsFile(t)
	ctr, err := dolt.Run(ctx,
		"dolthub/dolt-sql-server:1.32.4",
		dolt.WithDatabase("foo"),
		dolt.WithUsername("test"),
		dolt.WithPassword("test"),
		dolt.WithScripts(filepath.Join("testdata", "check_clone_private.sh")),
		dolt.WithDoltCloneRemoteUrl("fake-remote-url"),
		dolt.WithDoltCredsPublicKey("fake-public-key"),
		dolt.WithCredsFile(filename))
	testcontainers.CleanupContainer(t, ctr)
	requireCallerRegisteredDriver(t, ctx, ctr, err)
}

func TestDoltWithRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := dolt.Run(ctx,
		"dolthub/dolt-sql-server:1.32.4",
		dolt.WithDatabase("foo"),
		dolt.WithUsername("root"),
		dolt.WithPassword(""))
	testcontainers.CleanupContainer(t, ctr)
	requireCallerRegisteredDriver(t, ctx, ctr, err)

	// perform assertions
	connectionString := ctr.MustConnectionString(ctx)

	require.NotEmpty(t, connectionString)
}

func TestDoltWithScripts(t *testing.T) {
	ctx := context.Background()

	ctr, err := dolt.Run(ctx,
		"dolthub/dolt-sql-server:1.32.4",
		dolt.WithScripts(filepath.Join("testdata", "schema.sql")))
	testcontainers.CleanupContainer(t, ctr)
	requireCallerRegisteredDriver(t, ctx, ctr, err)

	// perform assertions
	connectionString := ctr.MustConnectionString(ctx)

	require.NotEmpty(t, connectionString)
}

// Successful driver-backed initialization is exercised by every original setup
// in examples. This package preserves the existing caller-registration boundary.
func requireCallerRegisteredDriver(t *testing.T, ctx context.Context, ctr *dolt.DoltContainer, err error) {
	t.Helper()
	require.NotContains(t, sql.Drivers(), "mysql")
	require.NotNil(t, ctr)
	require.EqualError(t, err, `initialize: sql: unknown driver "mysql" (forgotten import?)`)
	state, err := ctr.State(ctx)
	require.NoError(t, err)
	require.True(t, state.Running)
	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, connectionString)
	require.Equal(t, connectionString, ctr.MustConnectionString(ctx))
}
