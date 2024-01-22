package cockroachdb_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func TestCockroach_Ping(t *testing.T) {
	ctx := context.Background()

	inputs := []struct {
		name string
		opts []testcontainers.ContainerCustomizer
		conn string
	}{
		{
			name: "defaults",
			conn: "postgres://root@localhost:xxxxx/defaultdb?sslmode=disable",
		},
		{
			name: "database",
			opts: []testcontainers.ContainerCustomizer{
				cockroachdb.WithDatabase("test"),
			},
			conn: "postgres://root@localhost:xxxxx/test?sslmode=disable",
		},
		{
			name: "user",
			opts: []testcontainers.ContainerCustomizer{
				cockroachdb.WithUser("foo"),
			},
			conn: "postgres://foo@localhost:xxxxx/defaultdb?sslmode=disable",
		},
		{
			name: "user & password",
			opts: []testcontainers.ContainerCustomizer{
				cockroachdb.WithUser("foo"),
				cockroachdb.WithPassword("bar"),
			},
			conn: "postgres://foo:bar@localhost:xxxxx/defaultdb?sslmode=disable",
		},
	}

	for _, input := range inputs {
		t.Run(input.name, func(t *testing.T) {
			container, err := cockroachdb.RunContainer(ctx, input.opts...)
			require.NoError(t, err)

			t.Cleanup(func() {
				err := container.Terminate(ctx)
				require.NoError(t, err)
			})

			connStr := container.MustConnectionString(ctx)
			require.Equal(t, input.conn, removePort(t, connStr))

			conn, err := pgx.Connect(ctx, connStr)
			require.NoError(t, err)

			err = conn.Ping(ctx)
			require.NoError(t, err)
		})
	}
}

func removePort(t *testing.T, dsn string) string {
	t.Helper()

	u, err := url.Parse(dsn)
	require.NoError(t, err)

	return strings.Replace(dsn, ":"+u.Port(), ":xxxxx", 1)
}
