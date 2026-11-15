package cassandra_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
)

func TestCassandra(t *testing.T) {
	ctx := context.Background()

	ctr, err := cassandra.Run(ctx, "cassandra:4.1.3")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	connectionHost, err := ctr.ConnectionHost(ctx)
	// }
	require.NoError(t, err)

	require.NotEmpty(t, connectionHost)
}

func TestCassandraWithConfigFile(t *testing.T) {
	ctx := context.Background()

	ctr, err := cassandra.Run(ctx, "cassandra:4.1.3", cassandra.WithConfigFile(filepath.Join("testdata", "config.yaml")))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionHost, err := ctr.ConnectionHost(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, connectionHost)
}

func TestCassandraWithInitScripts(t *testing.T) {
	t.Run("with init cql script", func(t *testing.T) {
		ctx := context.Background()

		// withInitScripts {
		ctr, err := cassandra.Run(ctx, "cassandra:4.1.3", cassandra.WithInitScripts(filepath.Join("testdata", "init.cql")))
		// }
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		// connectionHost {
		connectionHost, err := ctr.ConnectionHost(ctx)
		// }
		require.NoError(t, err)

		require.NotEmpty(t, connectionHost)
	})

	t.Run("with init bash script", func(t *testing.T) {
		ctx := context.Background()

		ctr, err := cassandra.Run(ctx, "cassandra:4.1.3", cassandra.WithInitScripts(filepath.Join("testdata", "init.sh")))
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		connectionHost, err := ctr.ConnectionHost(ctx)
		require.NoError(t, err)

		require.NotEmpty(t, connectionHost)
	})
}

func TestCassandraWithTLS(t *testing.T) {
	ctx := context.Background()

	// withTLS {
	ctr, err := cassandra.Run(ctx, "cassandra:4.1.3", cassandra.WithTLS())
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// tlsConfig {
	// Verify TLS config is available
	tlsConfig, err := ctr.TLSConfig()
	require.NoError(t, err)
	// }
	require.NotNil(t, tlsConfig, "TLSConfig should not be nil when TLS is enabled")

	// Get SSL connection host
	// connectionHostSSL {
	connectionHost, err := ctr.ConnectionHost(ctx)
	// }
	require.NoError(t, err)

	require.NotEmpty(t, connectionHost)
}
