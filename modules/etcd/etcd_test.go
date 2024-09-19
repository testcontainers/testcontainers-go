package etcd_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/etcd"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	c, r, err := ctr.Exec(ctx, []string{"etcdctl", "member", "list"}, tcexec.Multiplexed())
	require.NoError(t, err)
	require.Equal(t, 0, c)

	output, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Contains(t, string(output), "default")
}

func TestRunCluster1Node(t *testing.T) {
	ctx := context.Background()

	ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithNodes("node1"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// the topology has only one node with no children
	require.Empty(t, ctr.Nodes)
	require.Equal(t, etcd.DefaultClusterToken, ctr.ClusterToken)
}

func TestRunClusterMultipleNodes(t *testing.T) {
	ctx := context.Background()

	ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithNodes("etcd-1", "etcd-2", "etcd-3"), etcd.WithClusterToken("My-cluster-t0k3n"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	require.Equal(t, "My-cluster-t0k3n", ctr.ClusterToken)

	// the topology has one parent node and two child nodes
	require.Len(t, ctr.Nodes, 2)

	for i, node := range ctr.Nodes {
		require.Empty(t, node.Nodes) // child nodes has no children

		c, r, err := node.Exec(ctx, []string{"etcdctl", "member", "list"}, tcexec.Multiplexed())
		require.NoError(t, err)

		output, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Contains(t, string(output), fmt.Sprintf("etcd-%d", i+1))

		require.Equal(t, 0, c)
		require.Equal(t, "My-cluster-t0k3n", node.ClusterToken)
	}
}

func TestRunClusterMultipleNodes_AutoTLS(t *testing.T) {
	ctx := context.Background()

	ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithNodes("etcd-1", "etcd-2", "etcd-3"), etcd.WithClusterToken("My-cluster-t0k3n"), etcd.WithAutoTLS())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	require.Equal(t, "My-cluster-t0k3n", ctr.ClusterToken)

	// the topology has one parent node and two child nodes
	require.Len(t, ctr.Nodes, 2)

	for i, node := range ctr.Nodes {
		require.Empty(t, node.Nodes) // child nodes has no children

		c, r, err := node.Exec(ctx, []string{"etcdctl", "member", "list"}, tcexec.Multiplexed())
		require.NoError(t, err)

		output, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Contains(t, string(output), fmt.Sprintf("etcd-%d", i+1))

		require.Equal(t, 0, c)
		require.Equal(t, "My-cluster-t0k3n", node.ClusterToken)
	}
}

func TestRun_PutGet(t *testing.T) {
	ctx := context.Background()

	ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithNodes("etcd-1", "etcd-2", "etcd-3"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   ctr.MustClientEndpoints(ctx),
		DialTimeout: 5 * time.Second,
	})
	require.NoError(t, err)
	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_, err = cli.Put(ctx, "sample_key", "sample_value")
	require.NoError(t, err)

	resp, err := cli.Get(ctx, "sample_key")
	require.NoError(t, err)

	require.Len(t, resp.Kvs, 1)
	require.Equal(t, "sample_value", string(resp.Kvs[0].Value))
}
