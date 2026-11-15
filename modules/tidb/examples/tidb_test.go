package tidb_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/tidb"
)

func TestTiDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := tidb.Run(ctx, "pingcap/tidb:v8.4.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	connectionString, err := ctr.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	mustConnectionString := ctr.MustConnectionString(ctx)
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

func TestTiDBWithDDLAndDML(t *testing.T) {
	ctx := context.Background()

	ctr, err := tidb.Run(ctx, "pingcap/tidb:v8.4.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sql.Open("mysql", connectionString)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS profile ( " +
		" id MEDIUMINT NOT NULL AUTO_INCREMENT, " +
		" name VARCHAR(30) NOT NULL, " +
		" PRIMARY KEY (id) " +
		")")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO profile (name) VALUES ('profile 1')")
	require.NoError(t, err)

	stmt, err := db.Prepare("SELECT name FROM profile")
	require.NoError(t, err)
	defer stmt.Close()

	row := stmt.QueryRow()
	var name string
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "profile 1", name)
}

func TestTiDBSelect1(t *testing.T) {
	ctx := context.Background()

	ctr, err := tidb.Run(ctx, "pingcap/tidb:v8.4.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sql.Open("mysql", connectionString)
	require.NoError(t, err)
	defer db.Close()

	row := db.QueryRow("SELECT 1")
	var result int
	err = row.Scan(&result)
	require.NoError(t, err)
	require.Equal(t, 1, result)
}
