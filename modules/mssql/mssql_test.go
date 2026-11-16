package mssql_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMSSQLServer(t *testing.T) {
	ctx := context.Background()

	ctr, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		mssql.WithAcceptEULA(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	assertConnectionString(t, connectionString)
	runSQL(t, ctx, ctr, "CREATE TABLE a_table ("+
		" [col_1] NVARCHAR(128) NOT NULL, "+
		" [col_2] NVARCHAR(128) NOT NULL, "+
		" PRIMARY KEY ([col_1], [col_2]) "+
		")")
}

func TestMSSQLServerWithMissingEulaOption(t *testing.T) {
	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		ctr, err := mssql.Run(ctx,
			"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
			testcontainers.WithAdditionalWaitStrategy(
				wait.ForLog("The SQL Server End-User License Agreement (EULA) must be accepted")),
		)
		testcontainers.CleanupContainer(t, ctr)
		require.Error(t, err)
	})

	t.Run("not-y", func(t *testing.T) {
		ctr, err := mssql.Run(ctx,
			"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
			testcontainers.WithEnv(map[string]string{"ACCEPT_EULA": "yes"}),
			testcontainers.WithAdditionalWaitStrategy(
				wait.ForLog("The SQL Server End-User License Agreement (EULA) must be accepted")),
		)
		testcontainers.CleanupContainer(t, ctr)
		require.Error(t, err)
	})
}

func TestMSSQLServerWithConnectionStringParameters(t *testing.T) {
	ctx := context.Background()

	ctr, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		mssql.WithAcceptEULA(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, err := ctr.ConnectionString(ctx, "encrypt=false", "TrustServerCertificate=true")
	require.NoError(t, err)

	parsed := assertConnectionString(t, connectionString)
	require.Equal(t, "false", parsed.Query().Get("encrypt"))
	require.Equal(t, "true", parsed.Query().Get("TrustServerCertificate"))
	runSQL(t, ctx, ctr, "CREATE TABLE a_table ("+
		" [col_1] NVARCHAR(128) NOT NULL, "+
		" [col_2] NVARCHAR(128) NOT NULL, "+
		" PRIMARY KEY ([col_1], [col_2]) "+
		")")
}

func TestMSSQLServerWithCustomStrongPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		mssql.WithAcceptEULA(),
		mssql.WithPassword("Strong@Passw0rd"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	parsed := assertConnectionString(t, connectionString)
	password, ok := parsed.User.Password()
	require.True(t, ok)
	require.Equal(t, "Strong@Passw0rd", password)
	require.JSONEq(t, `[{"connected":1}]`, runSQL(t, ctx, ctr, "SELECT 1 AS connected FOR JSON PATH"))
}

// tests that a weak password is not accepted by the container due to Microsoft's password strength policy
func TestMSSQLServerWithInvalidPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Password validation failed")),
		mssql.WithAcceptEULA(),
		mssql.WithPassword("weakPassword"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

//go:embed testdata/seed.sql
var seedSQLContent []byte

// tests that a container can be created with a DDL script
func TestMSSQLServerWithScriptsDDL(t *testing.T) {
	const password = "MyCustom@Passw0rd"

	// assertContainer contains the logic for asserting the test
	assertContainer := func(t *testing.T, ctx context.Context, image string, options ...testcontainers.ContainerCustomizer) {
		t.Helper()

		ctr, err := mssql.Run(ctx,
			image,
			append([]testcontainers.ContainerCustomizer{mssql.WithAcceptEULA()}, options...)...,
		)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		connectionString, err := ctr.ConnectionString(ctx)
		require.NoError(t, err)

		assertConnectionString(t, connectionString)
		result := runSQL(t, ctx, ctr, "SELECT ID AS id, ToppingName AS topping_name, Deliciousness AS deliciousness FROM pizza_palace.pizzas ORDER BY ID FOR JSON PATH")
		type pizza struct {
			ID            int    `json:"id"`
			ToppingName   string `json:"topping_name"`
			Deliciousness string `json:"deliciousness"`
		}
		var got []pizza
		require.NoError(t, json.Unmarshal([]byte(result), &got))
		require.Equal(t, []pizza{
			{1, "Pineapple", "Controversial but tasty"},
			{2, "Pepperoni", "Classic never fails"},
		}, got)
	}

	ctx := context.Background()

	t.Run("WithPassword/beforeWithScripts", func(t *testing.T) {
		assertContainer(t, ctx,
			"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
			mssql.WithPassword(password),
			mssql.WithInitSQL(bytes.NewReader(seedSQLContent)),
		)
	})

	t.Run("WithPassword/afterWithScripts", func(t *testing.T) {
		assertContainer(t, ctx,
			"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
			mssql.WithInitSQL(bytes.NewReader(seedSQLContent)),
			mssql.WithPassword(password),
		)
	})

	t.Run("2019-CU30-ubuntu-20.04/oldSQLCmd", func(t *testing.T) {
		assertContainer(t, ctx,
			"mcr.microsoft.com/mssql/server:2019-CU30-ubuntu-20.04",
			mssql.WithPassword(password),
			mssql.WithInitSQL(bytes.NewReader(seedSQLContent)),
		)
	})
}

// runSQL uses the official sqlcmd supplied by either supported image. -C trusts
// the image's self-signed certificate with both ODBC tool generations.
func runSQL(t *testing.T, ctx context.Context, ctr *mssql.MSSQLServerContainer, query string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	code, output, err := ctr.Exec(ctx, []string{
		"/bin/sh", "-c",
		`if [ -x /opt/mssql-tools18/bin/sqlcmd ]; then exec /opt/mssql-tools18/bin/sqlcmd "$@"; else exec /opt/mssql-tools/bin/sqlcmd "$@"; fi`,
		"sqlcmd", "-S", "localhost", "-U", "sa", "-P", ctr.Password(),
		"-C", "-b", "-l", "30", "-t", "30", "-h", "-1", "-w", "65535",
		"-Q", "SET NOCOUNT ON; " + query,
	}, exec.Multiplexed())
	require.NoError(t, err)
	const maxResponse = 1 << 20
	result, err := io.ReadAll(io.LimitReader(output, maxResponse+1))
	require.NoError(t, err)
	require.LessOrEqual(t, len(result), maxResponse)
	require.Zero(t, code, "%s", result)
	return strings.TrimSpace(string(result))
}

func assertConnectionString(t *testing.T, connectionString string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(connectionString)
	require.NoError(t, err)
	require.Equal(t, "sqlserver", parsed.Scheme)
	require.NotEmpty(t, parsed.Hostname())
	require.NotEmpty(t, parsed.Port())
	require.Equal(t, "sa", parsed.User.Username())
	return parsed
}
