package mariadb_test

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

func TestMariaDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := mariadb.Run(ctx, "mariadb:12.3.2")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	// By default, MariaDB transmits data between the server and clients without encrypting it.
	connectionString, err := ctr.ConnectionString(ctx, "tls=false")
	// }
	require.NoError(t, err)

	mustConnectionString := ctr.MustConnectionString(ctx, "tls=false")
	require.Equal(t, connectionString, mustConnectionString)
	require.True(t, strings.HasSuffix(connectionString, "?tls=false"))

	require.NotEmpty(t, connectionString)

	runSQL(t, ctx, ctr, "test", "test", "test", "CREATE TABLE IF NOT EXISTS a_table ( \n"+
		" `col_1` VARCHAR(128) NOT NULL, \n"+
		" `col_2` VARCHAR(128) NOT NULL, \n"+
		" PRIMARY KEY (`col_1`, `col_2`) \n"+
		")")
}

func TestMariaDBWithNonRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	_, err := mariadb.Run(ctx,
		"mariadb:12.3.2",
		mariadb.WithDatabase("foo"),
		mariadb.WithUsername("test"),
		mariadb.WithPassword(""))
	require.EqualError(t, err, "run mariadb: customize: empty password can be used only with the root user")
}

func TestMariaDBWithRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := mariadb.Run(ctx,
		"mariadb:12.3.2",
		mariadb.WithDatabase("foo"),
		mariadb.WithUsername("root"),
		mariadb.WithPassword(""))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, connectionString)

	runSQL(t, ctx, ctr, "root", "", "foo", "CREATE TABLE IF NOT EXISTS a_table ( \n"+
		" `col_1` VARCHAR(128) NOT NULL, \n"+
		" `col_2` VARCHAR(128) NOT NULL, \n"+
		" PRIMARY KEY (`col_1`, `col_2`) \n"+
		")")
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

	ctr, err := mariadb.Run(ctx, "mariadb:12.3.2",
		mariadb.WithConfigFile(filepath.Join("testdata", "my.cnf")))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, connectionString)

	// In MariaDB 10.2.2 and later, the default file format is Barracuda and Antelope is deprecated.
	// Barracuda is a newer InnoDB file format. It supports the COMPACT, REDUNDANT, DYNAMIC and
	// COMPRESSED row formats. Tables with large BLOB or TEXT columns in particular could benefit
	// from the dynamic row format.
	innodbFileFormat := runSQL(t, ctx, ctr, "test", "test", "test", "SELECT @@GLOBAL.innodb_default_row_format")
	require.Equal(t, "dynamic", innodbFileFormat)
}

func TestMariaDBWithScripts(t *testing.T) {
	ctx := context.Background()

	ctr, err := mariadb.Run(ctx,
		"mariadb:12.3.2",
		mariadb.WithScripts(filepath.Join("testdata", "schema.sql")))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	assertDataCanBeFetched(t, ctx, ctr)
}

func assertDataCanBeFetched(t *testing.T, ctx context.Context, container *mariadb.MariaDBContainer) {
	t.Helper()
	connectionString, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, connectionString)

	name := runSQL(t, ctx, container, "test", "test", "test", "SELECT name from profile")
	require.Equal(t, "profile 1", name)
}

// runSQL uses the official in-image client over TCP. The examples module separately
// verifies the generated public DSN with the Go driver, including TLS options.
func runSQL(t *testing.T, ctx context.Context, container *mariadb.MariaDBContainer, user, password, database, query string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	command := []string{
		"sh", "-c", `if command -v mariadb >/dev/null 2>&1; then exec mariadb "$@"; else exec mysql "$@"; fi`,
		"mariadb-client", "--skip-ssl",
		"--protocol=TCP", "--host=127.0.0.1", "--port=3306",
		"--connect-timeout=10", "--user=" + user, "--database=" + database,
		"--batch", "--skip-column-names", "--execute=" + query,
	}
	exitCode, reader, err := container.Exec(ctx, command, exec.WithEnv([]string{"MYSQL_PWD=" + password}), exec.Multiplexed())
	require.NoError(t, err)
	output, err := io.ReadAll(io.LimitReader(reader, 1024*1024+1))
	require.NoError(t, err)
	require.LessOrEqual(t, len(output), 1024*1024, "SQL client output exceeded limit")
	require.Zero(t, exitCode, "SQL client failed: %s", output)
	return strings.TrimSpace(string(output))
}
