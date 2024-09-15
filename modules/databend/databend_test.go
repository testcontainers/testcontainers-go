package databend_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/databend"
)

func TestDatabend(t *testing.T) {
	ctx := context.Background()

	ctr, err := databend.Run(ctx, "datafuselabs/databend:v1.2.615")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	// connectionString {
	connectionString, err := ctr.ConnectionString(ctx, "sslmode=disable")
	// }
	require.NoError(t, err)

	mustConnectionString := ctr.MustConnectionString(ctx, "sslmode=disable")
	require.Equal(t, connectionString, mustConnectionString)

	db, err := sql.Open("databend", connectionString)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS a_table ( \n" +
		" `col_1` VARCHAR(128) NOT NULL, \n" +
		" `col_2` VARCHAR(128) NOT NULL, \n" +
		")")
	require.NoError(t, err)
}

func TestDatabendWithNonRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := databend.Run(ctx,
		"datafuselabs/databend:v1.2.615",
		databend.WithUsername("databend"),
		databend.WithPassword("databend"))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}

func TestDatabendWithRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := databend.Run(ctx,
		"datafuselabs/databend:v1.2.615",
		databend.WithUsername("databend"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sql.Open("databend", connectionString)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS a_table ( \n" +
		" `col_1` VARCHAR(128) NOT NULL, \n" +
		" `col_2` VARCHAR(128) NOT NULL, \n" +
		")")
	require.NoError(t, err)
}
