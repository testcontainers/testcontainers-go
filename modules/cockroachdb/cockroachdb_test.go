package cockroachdb_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

const testImage = "cockroachdb/cockroach:latest-v23.1"

func TestRun(t *testing.T) {
	testContainer(t)
}

func TestRun_WithAllOptions(t *testing.T) {
	testContainer(t,
		cockroachdb.WithDatabase("testDatabase"),
		cockroachdb.WithStoreSize("50%"),
		cockroachdb.WithUser("testUser"),
		cockroachdb.WithPassword("testPassword"),
		cockroachdb.WithNoClusterDefaults(),
		cockroachdb.WithInitScripts("testdata/__init.sql"),
		// WithInsecure is not present as it is incompatible with WithPassword.
	)
}

func TestRun_WithInsecure(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		testContainer(t, cockroachdb.WithInsecure())
	})

	t.Run("invalid-password-insecure", func(t *testing.T) {
		_, err := cockroachdb.Run(context.Background(), testImage,
			cockroachdb.WithPassword("testPassword"),
			cockroachdb.WithInsecure(),
		)
		require.Error(t, err)
	})

	t.Run("invalid-insecure-password", func(t *testing.T) {
		_, err := cockroachdb.Run(context.Background(), testImage,
			cockroachdb.WithInsecure(),
			cockroachdb.WithPassword("testPassword"),
		)
		require.Error(t, err)
	})
}

// testContainer runs a CockroachDB container and validates its functionality.
func testContainer(t *testing.T, opts ...testcontainers.ContainerCustomizer) {
	t.Helper()

	ctx := context.Background()
	ctr, err := cockroachdb.Run(ctx, testImage, opts...)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.NotNil(t, ctr)

	// Check a raw connection with a ping.
	cfg, err := ctr.ConnectionConfig(ctx)
	require.NoError(t, err)

	conn, err := pgx.ConnectConfig(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.Close(ctx))
	})

	err = conn.Ping(ctx)
	require.NoError(t, err)

	// Check an SQL connection with a queries.
	addr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sql.Open("pgx/v5", addr)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "CREATE TABLE test (id INT PRIMARY KEY)")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO test (id) VALUES (523123)")
	require.NoError(t, err)

	var id int
	err = db.QueryRowContext(ctx, "SELECT id FROM test").Scan(&id)
	require.NoError(t, err)
	require.Equal(t, 523123, id)
}
