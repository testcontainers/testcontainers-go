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

	"github.com/testcontainers/testcontainers-go/modules/nebulagraph"
)

func TestNebulaGraphContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	container, err := nebulagraph.RunContainer(ctx)
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
	t.Run("basic operations", func(t *testing.T) {
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

		// Test SHOW HOSTS query
		result, err := g.Execute(ctx, *auth.SessionID, []byte("SHOW HOSTS;"))
		require.NoError(t, err)
		require.Equal(t, nebula.ErrorCode_SUCCEEDED, result.GetErrorCode(), "Query error: %s", result.GetErrorMsg())

		// Validate result contains our storage node
		resultSet, err := nebula_sirius.GenResultSet(result)
		require.NoError(t, err)

		// Convert result to string for validation
		rows := resultSet.GetRows()
		require.NotEmpty(t, rows, "Expected at least one row in SHOW HOSTS output")

		// Check the host status in the result
		hasStoraged := false
		for _, row := range rows {
			vals := row.GetValues()
			if len(vals) > 0 {
				hostVal := vals[0]
				statusVal := vals[2] // Status is typically the 3rd column
				if hostVal != nil && statusVal != nil {
					if string(hostVal.GetSVal()) == "storaged0" && string(statusVal.GetSVal()) == "ONLINE" {
						hasStoraged = true
						break
					}
				}
			}
		}
		require.True(t, hasStoraged, "Expected to find storaged0 in ONLINE state")
	})
}
