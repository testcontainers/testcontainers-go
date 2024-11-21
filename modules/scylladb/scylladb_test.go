package scylladb_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modules/scylladb"
)

func TestScylla(t *testing.T) {
	ctx := context.Background()

	ctr, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithShardAwareness(),
	)
	require.NoError(t, err)

	t.Run("test without shard awareness", func(t *testing.T) {
		host, err := ctr.ConnectionHost(ctx)
		require.NoError(t, err)

		cluster := gocql.NewCluster(host)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var address string
		err = session.Query("SELECT address FROM system.clients").Scan(&address)
		require.NoError(t, err)
	})

	t.Run("test with shard awareness", func(t *testing.T) {
		host, err := ctr.ShardAwareConnectionHost(ctx)
		require.NoError(t, err)

		cluster := gocql.NewCluster(host)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var address string
		err = session.Query("SELECT address FROM system.clients").Scan(&address)
		require.NoError(t, err)
	})
}

func TestScyllaWithConfigFile(t *testing.T) {
	ctx := context.Background()

	ctr, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithConfigFile(filepath.Join("testdata", "scylla.yaml")),
		scylladb.WithShardAwareness(),
	)
	require.NoError(t, err)

	t.Run("test without shard awareness", func(t *testing.T) {
		host, err := ctr.ConnectionHost(ctx)
		require.NoError(t, err)

		cluster := gocql.NewCluster(host)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var cluster_name string
		err = session.Query("SELECT cluster_name FROM system.local").Scan(&cluster_name)
		require.NoError(t, err)

		require.Equal(t, "Amazing ScyllaDB Test", cluster_name)
	})

	t.Run("test with shard awareness", func(t *testing.T) {
		host, err := ctr.ShardAwareConnectionHost(ctx)
		require.NoError(t, err)

		cluster := gocql.NewCluster(host)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var cluster_name string
		err = session.Query("SELECT cluster_name FROM system.local").Scan(&cluster_name)
		require.NoError(t, err)

		require.Equal(t, "Amazing ScyllaDB Test", cluster_name)
	})
}
