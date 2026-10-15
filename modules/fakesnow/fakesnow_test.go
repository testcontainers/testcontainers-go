package fakesnow_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/snowflakedb/gosnowflake"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/fakesnow"
)

const image = "ghcr.io/tekumara/fakesnow:0.11.21"

// open returns a database handle for ctr, with the default database and schema created.
func open(ctx context.Context, t *testing.T, ctr *fakesnow.Container) *sql.DB {
	t.Helper()

	dsn, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sql.Open("snowflake", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS db1")
	require.NoError(t, err)
	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS db1.schema1")
	require.NoError(t, err)

	return db
}

func TestRun(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakesnow.Run(ctx, image)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	db := open(ctx, t, ctr)
	require.NoError(t, db.Ping())

	var n int
	require.NoError(t, db.QueryRow("SELECT 42").Scan(&n))
	require.Equal(t, 42, n)
}

// The point of fakesnow over a plain database is the Snowflake-flavoured SQL,
// so check the parts a Snowflake user would reach for.
func TestSnowflakeSQL(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakesnow.Run(ctx, image)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	db := open(ctx, t, ctr)

	t.Run("variant", func(t *testing.T) {
		var got int
		err := db.QueryRow(`SELECT PARSE_JSON('{"k":[1,2]}'):k[1]`).Scan(&got)
		require.NoError(t, err)
		require.Equal(t, 2, got)
	})

	t.Run("lateral flatten", func(t *testing.T) {
		var got int
		err := db.QueryRow(`SELECT count(*) FROM LATERAL FLATTEN(input => PARSE_JSON('[1,2,3]'))`).Scan(&got)
		require.NoError(t, err)
		require.Equal(t, 3, got)
	})

	t.Run("qualify", func(t *testing.T) {
		_, err := db.Exec("CREATE OR REPLACE TABLE qualify_test (id int)")
		require.NoError(t, err)
		_, err = db.Exec("INSERT INTO qualify_test VALUES (1), (2)")
		require.NoError(t, err)

		var got int
		err = db.QueryRow(`SELECT count(*) FROM (
			SELECT id, ROW_NUMBER() OVER (ORDER BY id) rn FROM qualify_test QUALIFY rn = 1
		)`).Scan(&got)
		require.NoError(t, err)
		require.Equal(t, 1, got)
	})
}

func TestTransaction(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakesnow.Run(ctx, image)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	db := open(ctx, t, ctr)

	_, err = db.Exec("CREATE OR REPLACE TABLE tx_test (i int)")
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)
	_, err = tx.Exec("INSERT INTO tx_test VALUES (1)")
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())

	var count int
	require.NoError(t, db.QueryRow("SELECT count(*) FROM tx_test").Scan(&count))
	require.Zero(t, count)
}

func TestRunWithOptions(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakesnow.Run(ctx, image,
		fakesnow.WithAccount("acme"),
		fakesnow.WithDatabase("custom_db"),
		fakesnow.WithSchema("custom_schema"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	dsn, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.Contains(t, dsn, "acme")
	require.Contains(t, dsn, "custom_db")
	require.Contains(t, dsn, "custom_schema")

	db, err := sql.Open("snowflake", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS custom_db")
	require.NoError(t, err)
	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS custom_db.custom_schema")
	require.NoError(t, err)

	var got string
	require.NoError(t, db.QueryRow("SELECT CURRENT_DATABASE()").Scan(&got))
	require.Equal(t, "CUSTOM_DB", got)
}

func TestWithAccountRejectsEmpty(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakesnow.Run(ctx, image, fakesnow.WithAccount(""))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}
