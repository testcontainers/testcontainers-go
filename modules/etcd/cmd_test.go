package etcd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_configureCMD(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		got := configureCMD(options{})
		want := []string{"etcd", "--name=default", "--listen-client-urls=http://0.0.0.0:2379", "--advertise-client-urls=http://0.0.0.0:2379"}
		require.Equal(t, want, got)
	})

	t.Run("with-node", func(t *testing.T) {
		got := configureCMD(options{
			nodeNames: []string{"node1"},
		})
		want := []string{
			"etcd",
			"--name=node1",
			"--initial-advertise-peer-urls=http://node1:2380",
			"--advertise-client-urls=http://node1:2379",
			"--listen-peer-urls=http://0.0.0.0:2380",
			"--listen-client-urls=http://0.0.0.0:2379",
			"--initial-cluster-state=new",
			"--initial-cluster=node1=http://node1:2380",
		}
		require.Equal(t, want, got)
	})

	t.Run("with-node-datadir", func(t *testing.T) {
		got := configureCMD(options{
			nodeNames:    []string{"node1"},
			mountDataDir: true,
		})
		want := []string{
			"etcd",
			"--name=node1",
			"--initial-advertise-peer-urls=http://node1:2380",
			"--advertise-client-urls=http://node1:2379",
			"--listen-peer-urls=http://0.0.0.0:2380",
			"--listen-client-urls=http://0.0.0.0:2379",
			"--initial-cluster-state=new",
			"--initial-cluster=node1=http://node1:2380",
			"--data-dir=/data.etcd",
		}
		require.Equal(t, want, got)
	})

	t.Run("with-node-datadir-additional-args", func(t *testing.T) {
		got := configureCMD(options{
			nodeNames:      []string{"node1"},
			mountDataDir:   true,
			additionalArgs: []string{"--auto-compaction-retention=1"},
		})
		want := []string{
			"etcd",
			"--name=node1",
			"--initial-advertise-peer-urls=http://node1:2380",
			"--advertise-client-urls=http://node1:2379",
			"--listen-peer-urls=http://0.0.0.0:2380",
			"--listen-client-urls=http://0.0.0.0:2379",
			"--initial-cluster-state=new",
			"--initial-cluster=node1=http://node1:2380",
			"--data-dir=/data.etcd",
			"--auto-compaction-retention=1",
		}
		require.Equal(t, want, got)
	})

	t.Run("with-cluster", func(t *testing.T) {
		got := configureCMD(options{
			nodeNames: []string{"node1", "node2"},
		})
		want := []string{
			"etcd",
			"--name=node1",
			"--initial-advertise-peer-urls=http://node1:2380",
			"--advertise-client-urls=http://node1:2379",
			"--listen-peer-urls=http://0.0.0.0:2380",
			"--listen-client-urls=http://0.0.0.0:2379",
			"--initial-cluster-state=new",
			"--initial-cluster=node1=http://node1:2380,node2=http://node2:2380",
		}
		require.Equal(t, want, got)
	})

	t.Run("with-cluster-token", func(t *testing.T) {
		got := configureCMD(options{
			nodeNames:    []string{"node1", "node2"},
			clusterToken: "token",
		})
		want := []string{
			"etcd",
			"--name=node1",
			"--initial-advertise-peer-urls=http://node1:2380",
			"--advertise-client-urls=http://node1:2379",
			"--listen-peer-urls=http://0.0.0.0:2380",
			"--listen-client-urls=http://0.0.0.0:2379",
			"--initial-cluster-state=new",
			"--initial-cluster=node1=http://node1:2380,node2=http://node2:2380",
			"--initial-cluster-token=token",
		}
		require.Equal(t, want, got)
	})
}
