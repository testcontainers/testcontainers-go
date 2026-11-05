package tidb_test

import (
	"context"
	"net"
	"testing"

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

	assertConnectionString(t, ctx, ctr, connectionString)
}

// TestTiDBWithDDLAndDML checks the endpoint used by the SQL test in examples/tidb_test.go.
func TestTiDBWithDDLAndDML(t *testing.T) {
	ctx := context.Background()

	ctr, err := tidb.Run(ctx, "pingcap/tidb:v8.4.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	assertConnectionString(t, ctx, ctr, connectionString)
}

// TestTiDBSelect1 checks the endpoint used by the SQL test in examples/tidb_test.go.
func TestTiDBSelect1(t *testing.T) {
	ctx := context.Background()

	ctr, err := tidb.Run(ctx, "pingcap/tidb:v8.4.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	assertConnectionString(t, ctx, ctr, connectionString)
}

func assertConnectionString(t *testing.T, ctx context.Context, ctr *tidb.Container, connectionString string) {
	t.Helper()

	endpoint, err := ctr.PortEndpoint(ctx, "4000/tcp", "")
	require.NoError(t, err)
	host, port, err := net.SplitHostPort(endpoint)
	require.NoError(t, err)
	require.NotEmpty(t, host)
	require.NotEmpty(t, port)
	require.Equal(t, "root:@tcp("+endpoint+")/test", connectionString)
}
