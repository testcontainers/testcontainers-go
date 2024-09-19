package mariadb_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	// Import mysql into the scope of this package (required)
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

func TestMariaDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := mariadb.Run(ctx, "mariadb:11.0.3")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	// By default, MariaDB transmits data between the server and clients without encrypting it.
	connectionString, err := ctr.ConnectionString(ctx, "tls=false")
	// }
	require.NoError(t, err)

	mustConnectionString := ctr.MustConnectionString(ctx, "tls=false")
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

func TestMariaDBWithNonRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	_, err := mariadb.Run(ctx,
		"mariadb:11.0.3",
		mariadb.WithDatabase("foo"),
		mariadb.WithUsername("test"),
		mariadb.WithPassword(""))
	if err.Error() != "empty password can be used only with the root user" {
		t.Fatal(err)
	}
}

func TestMariaDBWithRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := mariadb.Run(ctx,
		"mariadb:11.0.3",
		mariadb.WithDatabase("foo"),
		mariadb.WithUsername("root"),
		mariadb.WithPassword(""))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

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

func TestMariaDBWithMySQLEnvVars(t *testing.T) {
	ctx := context.Background()

	ctr, err := mariadb.Run(ctx, "mariadb:10.3.29",
		mariadb.WithScripts(filepath.Join("testdata", "schema.sql")))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	assertDataCanBeFetched(t, ctx, ctr)
}

func TestMariaDBWithConfigFile(t *testing.T) {
	ctx := context.Background()

	ctr, err := mariadb.Run(ctx, "mariadb:11.0.3",
		mariadb.WithConfigFile(filepath.Join("testdata", "my.cnf")))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sql.Open("mysql", connectionString)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	// In MariaDB 10.2.2 and later, the default file format is Barracuda and Antelope is deprecated.
	// Barracuda is a newer InnoDB file format. It supports the COMPACT, REDUNDANT, DYNAMIC and
	// COMPRESSED row formats. Tables with large BLOB or TEXT columns in particular could benefit
	// from the dynamic row format.
	stmt, err := db.Prepare("SELECT @@GLOBAL.innodb_default_row_format")
	require.NoError(t, err)

	defer stmt.Close()
	row := stmt.QueryRow()
	innodbFileFormat := ""
	err = row.Scan(&innodbFileFormat)
	require.NoError(t, err)
	require.Equal(t, "dynamic", innodbFileFormat)
}

func TestMariaDBWithScripts(t *testing.T) {
	ctx := context.Background()

	ctr, err := mariadb.Run(ctx,
		"mariadb:11.0.3",
		mariadb.WithScripts(filepath.Join("testdata", "schema.sql")))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	assertDataCanBeFetched(t, ctx, ctr)
}

func assertDataCanBeFetched(t *testing.T, ctx context.Context, container *mariadb.MariaDBContainer) {
	connectionString, err := container.ConnectionString(ctx)
	require.NoError(t, err)
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
