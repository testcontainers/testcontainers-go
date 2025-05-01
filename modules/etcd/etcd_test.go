package etcd_test

import (
	"context"
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
	require.Zero(t, c)

	output, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Contains(t, string(output), "default")
}

func TestRun_PutGet(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name    string
		builder func(t *testing.T) *etcd.EtcdContainer
	}{
		{
			name: "single_node",
			builder: func(t *testing.T) *etcd.EtcdContainer {
				ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14")
				require.NoError(t, err)
				return ctr
			},
		},
		{
			name: "multiple_nodes",
			builder: func(t *testing.T) *etcd.EtcdContainer {
				ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithNodes("etcd-1", "etcd-2", "etcd-3"))
				require.NoError(t, err)
				return ctr
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctr := tc.builder(t)
			require.NoError(t, testcontainers.TerminateContainer(ctr))

			clientEndpoints, err := ctr.ClientEndpoints(ctx)
			require.NoError(t, err)

			cli, err := clientv3.New(clientv3.Config{
				Endpoints:   clientEndpoints,
				DialTimeout: 5 * time.Second,
			})
			require.NoError(t, err)
			defer func(cli *clientv3.Client) {
				require.NoError(t, cli.Close())
			}(cli)

			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			_, err = cli.Put(ctx, "sample_key", "sample_value")
			require.NoError(t, err)

			resp, err := cli.Get(ctx, "sample_key")
			require.NoError(t, err)

			require.Len(t, resp.Kvs, 1)
			require.Equal(t, "sample_value", string(resp.Kvs[0].Value))
		})

	}
}
