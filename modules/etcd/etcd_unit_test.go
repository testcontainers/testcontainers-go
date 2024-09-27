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

	ctr, err := Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// the topology has only one node with no children
	require.Empty(t, ctr.childNodes)
	require.Equal(t, defaultClusterToken, ctr.opts.clusterToken)
}

func TestRunClusterMultipleNodes(t *testing.T) {
	testCases := []struct {
		name  string
		node1 string
		node2 string
		nodes []string
	}{
		{
			name:  "2-nodes",
			node1: "etcd-1",
			node2: "etcd-2",
			nodes: []string{},
		},
		{
			name:  "3-nodes",
			node1: "etcd-1",
			node2: "etcd-2",
			nodes: []string{"etcd-3"},
		},
	}

	const clusterToken string = "My-cluster-t0k3n"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ctr, err := Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", WithNodes(tc.node1, tc.node2, tc.nodes...), WithClusterToken(clusterToken))
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			require.Equal(t, clusterToken, ctr.opts.clusterToken)

			// the topology has one parent node, one child node and optionally more child nodes
			// depending on the number of nodes specified
			require.Len(t, ctr.childNodes, 1+len(tc.nodes))

			for i, node := range ctr.childNodes {
				require.Empty(t, node.childNodes) // child nodes has no children

				c, r, err := node.Exec(ctx, []string{"etcdctl", "member", "list"}, tcexec.Multiplexed())
				require.NoError(t, err)

				output, err := io.ReadAll(r)
				require.NoError(t, err)
				require.Contains(t, string(output), fmt.Sprintf("etcd-%d", i+1))

				require.Zero(t, c)
				require.Equal(t, clusterToken, node.opts.clusterToken)
			}
		})
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

	_, err = cli.ContainerInspect(context.Background(), ctr.GetContainerID())
	require.True(t, errdefs.IsNotFound(err))

	for _, child := range ctr.childNodes {
		_, err := cli.ContainerInspect(context.Background(), child.GetContainerID())
		require.True(t, errdefs.IsNotFound(err))
	}

	_, err = cli.NetworkInspect(context.Background(), ctr.opts.clusterNetwork.ID, network.InspectOptions{})
	require.True(t, errdefs.IsNotFound(err))
}
