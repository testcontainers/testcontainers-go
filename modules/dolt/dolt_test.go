package dolt_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	// Import mysql into the scope of this package (required)
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dolt"
)

func TestDolt(t *testing.T) {
	ctx := context.Background()

	ctr, err := dolt.Run(ctx, "dolthub/dolt-sql-server:1.32.4")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	// connectionString {
	connectionString, err := ctr.ConnectionString(ctx)
	// }
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
	require.NoError(t, err)
}

func createTestCredsFile(t *testing.T) string {
	file, err := os.CreateTemp(t.TempDir(), "prefix")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	_, err = file.WriteString("some-fake-creds")
	if err != nil {
		t.Fatal(err)
	}
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
	require.NoError(t, err)
}

func TestDoltWithRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := dolt.Run(ctx,
		"dolthub/dolt-sql-server:1.32.4",
		dolt.WithDatabase("foo"),
		dolt.WithUsername("root"),
		dolt.WithPassword(""))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString := ctr.MustConnectionString(ctx)

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

func TestDoltWithScripts(t *testing.T) {
	ctx := context.Background()

	ctr, err := dolt.Run(ctx,
		"dolthub/dolt-sql-server:1.32.4",
		dolt.WithScripts(filepath.Join("testdata", "schema.sql")))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString := ctr.MustConnectionString(ctx)

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
