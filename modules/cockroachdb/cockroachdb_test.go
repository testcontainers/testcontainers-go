package cockroachdb_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func TestCockroach(t *testing.T) {
	ctx := context.Background()

	t.Run("ping default database", func(t *testing.T) {
		container, err := cockroachdb.RunContainer(ctx)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := container.Terminate(ctx)
			require.NoError(t, err)
		})

		conn, err := pgx.Connect(ctx, container.MustConnectionString(ctx))
		require.NoError(t, err)

		err = conn.Ping(ctx)
		require.NoError(t, err)
	})

	t.Run("ping custom database", func(t *testing.T) {
		container, err := cockroachdb.RunContainer(ctx, cockroachdb.WithDatabase("test"))
		require.NoError(t, err)

		t.Cleanup(func() {
			err := container.Terminate(ctx)
			require.NoError(t, err)
		})

		dsn, err := container.ConnectionString(ctx)
		require.NoError(t, err)

		u, err := url.Parse(dsn)
		require.NoError(t, err)
		require.Equal(t, "/test", u.Path)

		conn, err := pgx.Connect(ctx, dsn)
		require.NoError(t, err)

		err = conn.Ping(ctx)
		require.NoError(t, err)
	})
}
