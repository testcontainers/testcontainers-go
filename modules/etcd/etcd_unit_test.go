package etcd

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
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
	require.Empty(t, ctr.childNodes)
	require.Equal(t, defaultClusterToken, ctr.opts.clusterToken)
}

func TestRunClusterMultipleNodes(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", WithNodes("etcd-1", "etcd-2", "etcd-3"), WithClusterToken("My-cluster-t0k3n"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	require.Equal(t, "My-cluster-t0k3n", ctr.opts.clusterToken)

	// the topology has one parent node and two child nodes
	require.Len(t, ctr.childNodes, 2)

	for i, node := range ctr.childNodes {
		require.Empty(t, node.childNodes) // child nodes has no children

		c, r, err := node.Exec(ctx, []string{"etcdctl", "member", "list"}, tcexec.Multiplexed())
		require.NoError(t, err)

		output, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Contains(t, string(output), fmt.Sprintf("etcd-%d", i+1))

		require.Zero(t, c)
		require.Equal(t, "My-cluster-t0k3n", node.opts.clusterToken)
	}
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

	for _, child := range ctr.childNodes {
		_, err := cli.ContainerInspect(context.Background(), child.GetContainerID())
		require.True(t, errdefs.IsNotFound(err))
	}

	_, err = cli.NetworkInspect(context.Background(), ctr.opts.clusterNetwork.ID, network.InspectOptions{})
	require.True(t, errdefs.IsNotFound(err))
}
