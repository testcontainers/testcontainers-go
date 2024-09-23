package mysql_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	// Import mysql into the scope of this package (required)
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

func TestMySQL(t *testing.T) {
	ctx := context.Background()

	ctr, err := mysql.Run(ctx, "mysql:8.0.36")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	// connectionString {
	connectionString, err := ctr.ConnectionString(ctx, "tls=skip-verify")
	// }
	require.NoError(t, err)

	mustConnectionString := ctr.MustConnectionString(ctx, "tls=skip-verify")
	require.Equal(t, connectionString, mustConnectionString)

	db, err := sql.Open("mysql", connectionString)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS a_table ( \n" +
		" `col_1` VARCHAR(128) NOT NULL, \n" +
		" `col_2` VARCHAR(128) NOT NULL, \n" +
		" PRIMARY KEY (`col_1`, `col_2`) \n" +
		")")
	require.NoError(t, err)
}

func TestMySQLWithNonRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithDatabase("foo"),
		mysql.WithUsername("test"),
		mysql.WithPassword(""))
	testcontainers.CleanupContainer(t, ctr)
	require.EqualError(t, err, "empty password can be used only with the root user")
}

func TestMySQLWithRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithDatabase("foo"),
		mysql.WithUsername("root"),
		mysql.WithPassword(""))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, _ := ctr.ConnectionString(ctx)

	db, err := sql.Open("mysql", connectionString)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS a_table ( \n" +
		" `col_1` VARCHAR(128) NOT NULL, \n" +
		" `col_2` VARCHAR(128) NOT NULL, \n" +
		" PRIMARY KEY (`col_1`, `col_2`) \n" +
		")")
	require.NoError(t, err)
}

func TestMySQLWithScripts(t *testing.T) {
	ctx := context.Background()

	ctr, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithScripts(filepath.Join("testdata", "schema.sql")))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, _ := ctr.ConnectionString(ctx)

	db, err := sql.Open("mysql", connectionString)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	stmt, err := db.Prepare("SELECT name from profile")
	require.NoError(t, err)
	defer stmt.Close()

	row := stmt.QueryRow()
	var name string
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "profile 1", name)
}
