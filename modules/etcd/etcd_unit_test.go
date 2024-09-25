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

func TestRunClusterMultipleNodes(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", WithNodes("etcd-1", "etcd-2", "etcd-3"), WithClusterToken("My-cluster-t0k3n"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	require.Equal(t, "My-cluster-t0k3n", ctr.ClusterToken)

	// the topology has one parent node and two child nodes
	require.Equal(t, 2, ctr.NodesCount())

	for i, node := range ctr.nodes {
		require.Zero(t, node.NodesCount()) // child nodes has no children

		c, r, err := node.Exec(ctx, []string{"etcdctl", "member", "list"}, tcexec.Multiplexed())
		require.NoError(t, err)

		output, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Contains(t, string(output), fmt.Sprintf("etcd-%d", i+1))

		require.Equal(t, 0, c)
		require.Equal(t, "My-cluster-t0k3n", node.ClusterToken)
	}
}
