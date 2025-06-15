package etcd

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
)

func TestRunCluster1Node(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// the topology has only one node with no children
	require.Empty(t, ctr.childNodes)
	require.Equal(t, defaultClusterToken, ctr.opts.clusterToken)
}

func TestRunClusterMultipleNodes(t *testing.T) {
	t.Run("2-nodes", testCluster(t, "etcd-1", "etcd-2"))
	t.Run("3-nodes", testCluster(t, "etcd-1", "etcd-2", "etcd-3"))
}

func TestTerminate(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", WithNodes("etcd-1", "etcd-2", "etcd-3"))
	require.NoError(t, err)
	require.NoError(t, ctr.Terminate(ctx))

	// verify that the network and the containers does no longer exist

	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)
	defer cli.Close()

	_, err = cli.ContainerInspect(context.Background(), ctr.GetContainerID())
	require.True(t, errdefs.IsNotFound(err))

	for _, child := range ctr.childNodes {
		_, err := cli.ContainerInspect(context.Background(), child.GetContainerID())
		require.True(t, errdefs.IsNotFound(err))
	}

	_, err = cli.NetworkInspect(context.Background(), ctr.opts.clusterNetwork.ID, network.InspectOptions{})
	require.True(t, errdefs.IsNotFound(err))
}

func TestTerminate_partiallyInitialised(t *testing.T) {
	newNetwork, err := tcnetwork.New(context.Background())
	require.NoError(t, err)

	ctr := &EtcdContainer{
		opts: options{
			clusterNetwork: newNetwork,
		},
	}

	require.NoError(t, ctr.Terminate(context.Background()))

	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)
	defer cli.Close()

	_, err = cli.NetworkInspect(context.Background(), ctr.opts.clusterNetwork.ID, network.InspectOptions{})
	require.True(t, errdefs.IsNotFound(err))
}

// testCluster is a helper function to test the creation of an etcd cluster with the specified nodes.
func testCluster(t *testing.T, node1 string, node2 string, nodes ...string) func(t *testing.T) {
	t.Helper()

	return func(tt *testing.T) {
		const clusterToken string = "My-cluster-t0k3n"

		ctx := context.Background()

		ctr, err := Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", WithNodes(node1, node2, nodes...), WithClusterToken(clusterToken))
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(tt, err)

		require.Equal(tt, clusterToken, ctr.opts.clusterToken)

		// the topology has one parent node, one child node and optionally more child nodes
		// depending on the number of nodes specified
		require.Len(tt, ctr.childNodes, 1+len(nodes))

		for i, node := range ctr.childNodes {
			require.Empty(t, node.childNodes) // child nodes has no children

			c, r, err := node.Exec(ctx, []string{"etcdctl", "member", "list"}, tcexec.Multiplexed())
			require.NoError(tt, err)

			output, err := io.ReadAll(r)
			require.NoError(t, err)
			require.Contains(t, string(output), fmt.Sprintf("etcd-%d", i+1))

			require.Zero(tt, c)
			require.Equal(tt, clusterToken, node.opts.clusterToken)
		}
	}
}
