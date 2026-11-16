package mysql_test

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
	require.True(t, strings.HasSuffix(connectionString, "?tls=skip-verify"))

	require.NotEmpty(t, connectionString)

	runSQL(t, ctx, ctr, "test", "test", "test", "CREATE TABLE IF NOT EXISTS a_table ( \n"+
		" `col_1` VARCHAR(128) NOT NULL, \n"+
		" `col_2` VARCHAR(128) NOT NULL, \n"+
		" PRIMARY KEY (`col_1`, `col_2`) \n"+
		")")
}

func TestMySQLWithNonRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithDatabase("foo"),
		mysql.WithUsername("test"),
		mysql.WithPassword(""))
	testcontainers.CleanupContainer(t, ctr)
	require.ErrorContains(t, err, "empty password can be used only with the root user")
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
	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, connectionString)

	runSQL(t, ctx, ctr, "root", "", "foo", "CREATE TABLE IF NOT EXISTS a_table ( \n"+
		" `col_1` VARCHAR(128) NOT NULL, \n"+
		" `col_2` VARCHAR(128) NOT NULL, \n"+
		" PRIMARY KEY (`col_1`, `col_2`) \n"+
		")")
}

func TestMySQLWithScripts(t *testing.T) {
	ctx := context.Background()

	ctr, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithScripts(filepath.Join("testdata", "schema.sql")))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, connectionString)

	name := runSQL(t, ctx, ctr, "test", "test", "test", "SELECT name from profile")
	require.Equal(t, "profile 1", name)
}

// runSQL uses the official in-image client over TCP. The examples module separately
// verifies the generated public DSN with the Go driver, including TLS options.
func runSQL(t *testing.T, ctx context.Context, container *mysql.MySQLContainer, user, password, database, query string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	command := []string{
		"mysql", "--ssl-mode=REQUIRED", "--protocol=TCP", "--host=127.0.0.1", "--port=3306",
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
