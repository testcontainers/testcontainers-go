package databend_test

import (
	"context"
	"net/url"
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

	assertConnectionString(t, ctx, ctr, connectionString)
}

func TestDatabendWithDefaultUserAndPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := databend.Run(ctx,
		"datafuselabs/databend:v1.2.615",
		databend.WithUsername("databend"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	assertConnectionString(t, ctx, ctr, connectionString)
}

func TestDatabendWithEmptyCredentials(t *testing.T) {
	ctx := context.Background()

	ctr, err := databend.Run(ctx,
		"datafuselabs/databend:v1.2.615",
		databend.WithUsername(""),
		databend.WithPassword(""))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}

func assertConnectionString(t *testing.T, ctx context.Context, ctr *databend.DatabendContainer, connectionString string) {
	t.Helper()

	endpoint, err := ctr.PortEndpoint(ctx, "8000/tcp", "")
	require.NoError(t, err)
	dsn, err := url.Parse(connectionString)
	require.NoError(t, err)
	require.Equal(t, "databend", dsn.Scheme)
	require.Equal(t, endpoint, dsn.Host)
	require.NotEmpty(t, dsn.Hostname())
	require.NotEmpty(t, dsn.Port())
	require.NotNil(t, dsn.User)
	require.Equal(t, "databend", dsn.User.Username())
	password, ok := dsn.User.Password()
	require.True(t, ok)
	require.Equal(t, "databend", password)
	require.Equal(t, "/default", dsn.Path)
	require.Equal(t, "disable", dsn.Query().Get("sslmode"))
}
