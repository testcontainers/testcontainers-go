package etcd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigureCMD(t *testing.T) {
	tests := []struct {
		name     string
		settings options
		want     []string
	}{
		{
			name:     "default",
			settings: options{},
			want:     []string{"etcd", "--name=default"},
		},
		{
			name: "with-node",
			settings: options{
				Nodes: []string{"node1"},
			},
			want: []string{
				"etcd",
				"--name=node1",
				"--initial-advertise-peer-urls=http://node1:2380",
				"--advertise-client-urls=http://node1:2379",
				"--listen-peer-urls=http://0.0.0.0:2380",
				"--listen-client-urls=http://0.0.0.0:2379",
				"--initial-cluster-state=new",
				"--initial-cluster=node1=http://node1:2380",
			},
		},
		{
			name: "with-node-datadir",
			settings: options{
				Nodes:        []string{"node1"},
				MountDataDir: true,
			},
			want: []string{
				"etcd",
				"--name=node1",
				"--initial-advertise-peer-urls=http://node1:2380",
				"--advertise-client-urls=http://node1:2379",
				"--listen-peer-urls=http://0.0.0.0:2380",
				"--listen-client-urls=http://0.0.0.0:2379",
				"--initial-cluster-state=new",
				"--initial-cluster=node1=http://node1:2380",
				"--data-dir=/data.etcd",
			},
		},
		{
			name: "with-node-datadir-additional-args",
			settings: options{
				Nodes:          []string{"node1"},
				MountDataDir:   true,
				AdditionalArgs: []string{"--auto-compaction-retention=1"},
			},
			want: []string{
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
			},
		},
		{
			name: "with-cluster",
			settings: options{
				Nodes: []string{"node1", "node2"},
			},
			want: []string{
				"etcd",
				"--name=node1",
				"--initial-advertise-peer-urls=http://node1:2380",
				"--advertise-client-urls=http://node1:2379",
				"--listen-peer-urls=http://0.0.0.0:2380",
				"--listen-client-urls=http://0.0.0.0:2379",
				"--initial-cluster-state=new",
				"--initial-cluster=node1=http://node1:2380,node2=http://node2:2380",
			},
		},
		{
			name: "with-cluster-token",
			settings: options{
				Nodes:        []string{"node1", "node2"},
				ClusterToken: "token",
			},
			want: []string{
				"etcd",
				"--name=node1",
				"--initial-advertise-peer-urls=http://node1:2380",
				"--advertise-client-urls=http://node1:2379",
				"--listen-peer-urls=http://0.0.0.0:2380",
				"--listen-client-urls=http://0.0.0.0:2379",
				"--initial-cluster-state=new",
				"--initial-cluster=node1=http://node1:2380,node2=http://node2:2380",
				"--initial-cluster-token=token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := configureCMD(tt.settings)
			require.Equal(t, tt.want, got)
		})
	}
}
