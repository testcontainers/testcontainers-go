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

func TestPutGet(t *testing.T) {
	t.Run("single_node", func(t *testing.T) {
		ctr, err := etcd.Run(context.Background(), "gcr.io/etcd-development/etcd:v3.5.14")
		testPutGet(t, ctr, err)
	})
	t.Run("multiple_nodes", func(t *testing.T) {
		ctr, err := etcd.Run(context.Background(), "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithNodes("etcd-1", "etcd-2", "etcd-3"))
		testPutGet(t, ctr, err)
	})
}

func testPutGet(t *testing.T, ctr *etcd.EtcdContainer, err error) {
	t.Helper()

	testcontainers.CleanupContainer(t, ctr)

	require.NoError(t, err)

	ctx := context.Background()

	clientEndpoints, err := ctr.ClientEndpoints(ctx)
	require.NoError(t, err)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   clientEndpoints,
		DialTimeout: 5 * time.Second,
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, cli.Close())
	}()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_, err = cli.Put(ctx, "sample_key", "sample_value")
	require.NoError(t, err)

	resp, err := cli.Get(ctx, "sample_key")
	require.NoError(t, err)

	require.Len(t, resp.Kvs, 1)
	require.Equal(t, "sample_value", string(resp.Kvs[0].Value))
}
