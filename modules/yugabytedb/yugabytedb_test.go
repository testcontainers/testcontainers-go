package yugabytedb_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"testing"

	_ "github.com/lib/pq"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yugabyte/gocql"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/yugabytedb"
)

func TestYugabyteDB_YSQL(t *testing.T) {
	t.Run("Run", func(t *testing.T) {
		ctx := context.Background()

		ctr, err := yugabytedb.Run(ctx, "yugabytedb/yugabyte:2024.1.3.0-b105")
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		ctrHost, err := ctr.Host(ctx)
		require.NoError(t, err)

		ctrPort, err := ctr.MappedPort(ctx, network.MustParsePort("5433/tcp"))
		require.NoError(t, err)

		ysqlConnStr, err := ctr.YSQLConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("postgres://yugabyte:yugabyte@%s:%s/yugabyte?sslmode=disable", ctrHost, strconv.FormatUint(uint64(ctrPort.Num()), 10)), ysqlConnStr)

		db, err := sql.Open("postgres", ysqlConnStr)
		require.NoError(t, err)
		require.NotNil(t, db)

		err = db.Ping()
		require.NoError(t, err)
	})

	t.Run("custom-options", func(t *testing.T) {
		ctx := context.Background()
		ctr, err := yugabytedb.Run(ctx, "yugabytedb/yugabyte:2024.1.3.0-b105",
			yugabytedb.WithDatabaseName("custom-db"),
			yugabytedb.WithDatabaseUser("custom-user"),
			yugabytedb.WithDatabasePassword("custom-password"),
		)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		ctrHost, err := ctr.Host(ctx)
		require.NoError(t, err)

		ctrPort, err := ctr.MappedPort(ctx, network.MustParsePort("5433/tcp"))
		require.NoError(t, err)

		ysqlConnStr, err := ctr.YSQLConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("postgres://custom-user:custom-password@%s:%s/custom-db?sslmode=disable", ctrHost, strconv.FormatUint(uint64(ctrPort.Num()), 10)), ysqlConnStr)

		db, err := sql.Open("postgres", ysqlConnStr)
		require.NoError(t, err)
		require.NotNil(t, db)

		err = db.Ping()
		require.NoError(t, err)
	})
}

func TestYugabyteDB_YCQL(t *testing.T) {
	t.Run("Run", func(t *testing.T) {
		ctx := context.Background()

		ctr, err := yugabytedb.Run(ctx, "yugabytedb/yugabyte:2024.1.3.0-b105")
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		ctrHost, err := ctr.Host(ctx)
		require.NoError(t, err)

		ctrPort, err := ctr.MappedPort(ctx, network.MustParsePort("9042/tcp"))
		require.NoError(t, err)

		cluster := gocql.NewCluster(net.JoinHostPort(ctrHost, strconv.FormatUint(uint64(ctrPort.Num()), 10)))
		cluster.Keyspace = "yugabyte"
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: "yugabyte",
			Password: "yugabyte",
		}

		session, err := cluster.CreateSession()
		require.NoError(t, err)
		session.Close()
	})

	t.Run("custom-options", func(t *testing.T) {
		ctx := context.Background()

		ctr, err := yugabytedb.Run(ctx, "yugabytedb/yugabyte:2024.1.3.0-b105",
			yugabytedb.WithKeyspace("custom-keyspace"),
			yugabytedb.WithUser("custom-user"),
			yugabytedb.WithPassword("custom-password"),
		)

		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		ctrHost, err := ctr.Host(ctx)
		require.NoError(t, err)

		ctrPort, err := ctr.MappedPort(ctx, network.MustParsePort("9042/tcp"))
		require.NoError(t, err)

		cluster := gocql.NewCluster(net.JoinHostPort(ctrHost, strconv.FormatUint(uint64(ctrPort.Num()), 10)))
		cluster.Keyspace = "custom-keyspace"
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: "custom-user",
			Password: "custom-password",
		}

		session, err := cluster.CreateSession()
		require.NoError(t, err)
		session.Close()
	})
}
