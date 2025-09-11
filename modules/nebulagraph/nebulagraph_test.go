package nebulagraph_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	pool "github.com/jolestar/go-commons-pool"
	nebula_sirius "github.com/nebula-contrib/nebula-sirius"
	"github.com/nebula-contrib/nebula-sirius/nebula"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/nebulagraph"
)

const (
	defaultGraphdImage   = "vesoft/nebula-graphd:v3.8.0"
	defaultMetadImage    = "vesoft/nebula-metad:v3.8.0"
	defaultStoragedImage = "vesoft/nebula-storaged:v3.8.0"
)

func TestNebulaGraphContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	container, err := nebulagraph.RunCluster(ctx,
		defaultGraphdImage, []testcontainers.ContainerCustomizer{},
		defaultStoragedImage, []testcontainers.ContainerCustomizer{},
		defaultMetadImage, []testcontainers.ContainerCustomizer{},
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	conn, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, conn)

	// Parse the connection string to get host and port
	host, portt, err := net.SplitHostPort(conn)
	require.NoError(t, err)

	portInt, err := strconv.Atoi(portt)
	require.NoError(t, err)

	// Create client factory
	clientFactory := nebula_sirius.NewNebulaClientFactory(
		&nebula_sirius.NebulaClientConfig{
			HostAddress: nebula_sirius.HostAddress{
				Host: host,
				Port: portInt,
			},
		},
		nebula_sirius.DefaultLogger{},
		nebula_sirius.DefaultClientNameGenerator,
	)

	// Create client pool
	nebulaClientPool := pool.NewObjectPool(
		ctx,
		clientFactory,
		&pool.ObjectPoolConfig{
			MaxIdle:  5,
			MaxTotal: 10,
		},
	)

	// Test client connection and basic queries
	t.Run("basic-operations", func(t *testing.T) {
		// Get a client from the pool
		clientObj, err := nebulaClientPool.BorrowObject(ctx)
		require.NoError(t, err)
		defer func() {
			err := nebulaClientPool.ReturnObject(ctx, clientObj)
			require.NoError(t, err)
		}()

		client := clientObj.(*nebula_sirius.WrappedNebulaClient)
		require.NotNil(t, client)

		// Get graph client
		g, err := client.GraphClient()
		require.NoError(t, err)

		// Authenticate
		auth, err := g.Authenticate(ctx, []byte("root"), []byte("nebula"))
		require.NoError(t, err)
		require.Equal(t, nebula.ErrorCode_SUCCEEDED, auth.GetErrorCode(), "Auth error: %s", auth.GetErrorMsg())

		// Test YIELD query
		result, err := g.Execute(ctx, *auth.SessionID, []byte("YIELD 1;"))
		require.NoError(t, err)
		require.Equal(t, nebula.ErrorCode_SUCCEEDED, result.GetErrorCode(), "Query error: %s", result.GetErrorMsg())

		// Validate result contains our storage node
		resultSet, err := nebula_sirius.GenResultSet(result)
		require.NoError(t, err)

		// Convert result to string for validation
		rows := resultSet.GetRows()
		require.NotEmpty(t, rows, "Expected at least one row in YIELD output")

		row := rows[0]
		require.NotNil(t, row, "Row should not be nil")

		vals := row.GetValues()
		require.NotEmpty(t, vals, "Row values should not be empty")
		require.Equal(t, vals[0].GetIVal(), int64(1), "Expected first column to be 1")
	})
}
