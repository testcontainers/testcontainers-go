package etcd

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

func TestRunCluster1Node(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", WithNodes("node1"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// the topology has only one node with no children
	require.Len(t, ctr.nodes, 0)
	require.Equal(t, DefaultClusterToken, ctr.ClusterToken)
}

func TestRunClusterMultipleNodes(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", WithNodes("etcd-1", "etcd-2", "etcd-3"), WithClusterToken("My-cluster-t0k3n"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	require.Equal(t, "My-cluster-t0k3n", ctr.ClusterToken)

	// the topology has one parent node and two child nodes
	require.Len(t, ctr.nodes, 2)

	for i, node := range ctr.nodes {
		require.Len(t, ctr.nodes, 0) // child nodes has no children

		c, r, err := node.Exec(ctx, []string{"etcdctl", "member", "list"}, tcexec.Multiplexed())
		require.NoError(t, err)

		output, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Contains(t, string(output), fmt.Sprintf("etcd-%d", i+1))

		require.Equal(t, 0, c)
		require.Equal(t, "My-cluster-t0k3n", node.ClusterToken)
	}
}
