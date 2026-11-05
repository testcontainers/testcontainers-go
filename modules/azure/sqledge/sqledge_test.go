package sqledge_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/sqledge"
)

func TestAzureSQLEdge(t *testing.T) {
	ctx := context.Background()

	ctr, err := sqledge.Run(ctx,
		"mcr.microsoft.com/azure-sql-edge:1.0.7",
		sqledge.WithAcceptEULA(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	assertConnectionString(t, ctx, ctr, "Strong!Passw0rd")
}

func TestAzureSQLEdgeWithMissingEULA(t *testing.T) {
	ctx := context.Background()

	ctr, err := sqledge.Run(ctx, "mcr.microsoft.com/azure-sql-edge:1.0.7")
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
	require.ErrorContains(t, err, "EULA not accepted")
}

func TestAzureSQLEdgeWithCustomPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := sqledge.Run(ctx,
		"mcr.microsoft.com/azure-sql-edge:1.0.7",
		sqledge.WithAcceptEULA(),
		sqledge.WithPassword("Strong@Passw0rd"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	assertConnectionString(t, ctx, ctr, "Strong@Passw0rd")
}

func TestAzureSQLEdgeWithConnectionStringParameters(t *testing.T) {
	ctx := context.Background()

	ctr, err := sqledge.Run(ctx,
		"mcr.microsoft.com/azure-sql-edge:1.0.7",
		sqledge.WithAcceptEULA(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	u := assertConnectionString(t, ctx, ctr, "Strong!Passw0rd", "encrypt=false", "TrustServerCertificate=true")
	require.Equal(t, "false", u.Query().Get("encrypt"))
	require.Equal(t, "true", u.Query().Get("TrustServerCertificate"))
}

func assertConnectionString(t *testing.T, ctx context.Context, ctr *sqledge.Container, password string, parameters ...string) *url.URL {
	t.Helper()

	state, err := ctr.State(ctx)
	require.NoError(t, err)
	require.True(t, state.Running)

	connectionString, err := ctr.ConnectionString(ctx, parameters...)
	require.NoError(t, err)
	u, err := url.Parse(connectionString)
	require.NoError(t, err)
	require.Equal(t, "sqlserver", u.Scheme)
	require.Equal(t, "sa", u.User.Username())
	actualPassword, ok := u.User.Password()
	require.True(t, ok)
	require.Equal(t, password, actualPassword)
	require.Equal(t, "master", u.Query().Get("database"))

	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	port, err := ctr.MappedPort(ctx, "1433/tcp")
	require.NoError(t, err)
	require.Equal(t, host, u.Hostname())
	require.Equal(t, port.Port(), u.Port())
	return u
}
