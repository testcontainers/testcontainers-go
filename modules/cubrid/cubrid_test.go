package cubrid_test

import (
	"context"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/cubrid"
)

const image = "cubrid/cubrid:11.4"

func TestCUBRID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ctr, err := cubrid.Run(ctx, image)
	testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
	require.NoError(t, err)

	require.Equal(t, "'TESTUSER'", csql(t, ctx, ctr, "testuser", "testpass", "testdb", "SELECT CURRENT_USER"))
	require.Equal(t, "'iso88591'", csql(t, ctx, ctr, "testuser", "testpass", "testdb", "SELECT CHARSET('a')"))

	// connectionString {
	connStr, err := ctr.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	endpoint, err := ctr.PortEndpoint(ctx, "33000/tcp", "")
	require.NoError(t, err)
	require.Equal(t, "cubrid://testuser:testpass@"+endpoint+"/testdb", connStr)
}

func TestCUBRIDWithOptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ctr, err := cubrid.Run(ctx, image,
		cubrid.WithDatabase("mydb"),
		cubrid.WithUsername("app"),
		cubrid.WithPassword("p@ss:w/rd"),
	)
	testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
	require.NoError(t, err)

	require.Equal(t, "'APP'", csql(t, ctx, ctr, "app", "p@ss:w/rd", "mydb", "SELECT CURRENT_USER"))

	connStr, err := ctr.ConnectionString(ctx, "autocommit=false")
	require.NoError(t, err)

	u, err := url.Parse(connStr)
	require.NoError(t, err)
	password, ok := u.User.Password()
	require.True(t, ok)
	require.Equal(t, "cubrid", u.Scheme)
	require.Equal(t, "app", u.User.Username())
	require.Equal(t, "p@ss:w/rd", password)
	require.Equal(t, "/mydb", u.Path)
	require.Equal(t, "autocommit=false", u.RawQuery)
}

func TestCUBRIDLocale(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("with-locale", func(t *testing.T) {
		t.Parallel()

		ctr, err := cubrid.Run(ctx, image, cubrid.WithLocale("ko_KR.utf8"))
		testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
		require.NoError(t, err)

		require.Equal(t, "'utf8'", csql(t, ctx, ctr, "testuser", "testpass", "testdb", "SELECT CHARSET('a')"))
	})

	t.Run("with-env", func(t *testing.T) {
		t.Parallel()

		ctr, err := cubrid.Run(ctx, image, testcontainers.WithEnv(map[string]string{"CUBRID_LOCALE": "ko_KR.euckr"}))
		testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
		require.NoError(t, err)

		require.Equal(t, "'euckr'", csql(t, ctx, ctr, "testuser", "testpass", "testdb", "SELECT CHARSET('a')"))
	})
}

func TestCUBRIDRejectsInvalidSettings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name string
		opt  testcontainers.ContainerCustomizer
		err  string
	}{
		{name: "empty-database", opt: cubrid.WithDatabase(""), err: "database must not be empty"},
		{name: "empty-username", opt: cubrid.WithUsername(""), err: "username must not be empty"},
		{name: "dba", opt: cubrid.WithUsername("dba"), err: `username "dba" is reserved by CUBRID`},
		{name: "public", opt: cubrid.WithUsername("PUBLIC"), err: `username "PUBLIC" is reserved by CUBRID`},
		{name: "empty-locale", opt: cubrid.WithLocale(""), err: "locale must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctr, err := cubrid.Run(ctx, image, tt.opt)
			testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
			require.ErrorContains(t, err, tt.err)
		})
	}
}

func TestCUBRIDFailsFast(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name string
		opt  testcontainers.ContainerCustomizer
		err  string
	}{
		{name: "unsupported-locale", opt: cubrid.WithLocale("xx_XX.utf8"), err: "Couldn't create database."},
		{name: "user-not-created", opt: cubrid.WithPassword("it's"), err: "ERROR:"},
		{name: "container-exits", opt: testcontainers.WithCmd("bash", "-c", "sleep 2; exit 3"), err: "container exited with code 3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctr, err := cubrid.Run(ctx, image, tt.opt)
			testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
			require.ErrorContains(t, err, tt.err)
		})
	}
}

func csql(t *testing.T, ctx context.Context, ctr *cubrid.Container, user, password, database, query string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := []string{"csql", "-u", user, "-p", password, database, "-c", query}
	exitCode, reader, err := ctr.Exec(ctx, cmd, exec.WithUser("cubrid"), exec.Multiplexed())
	require.NoError(t, err)
	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Zero(t, exitCode, "csql failed: %s", output)

	lines := strings.Split(string(output), "\n")
	for i := 0; i+1 < len(lines); i++ {
		if sep := strings.TrimSpace(lines[i]); sep != "" && strings.Trim(sep, "=") == "" {
			return strings.TrimSpace(lines[i+1])
		}
	}
	t.Fatalf("no result in csql output: %s", output)
	return ""
}
