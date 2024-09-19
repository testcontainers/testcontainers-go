package etcd

import (
	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	currentNode    int
	clusterNetwork *testcontainers.DockerNetwork
	Nodes          []string
	ClusterToken   string
	AdditionalArgs []string
	MountDataDir   bool
}

func defaultOptions() options {
	return options{
		currentNode:    0,
		Nodes:          []string{},
		AdditionalArgs: []string{},
		ClusterToken:   DefaultClusterToken,
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Redpanda container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithAdditionalArgs is an option to pass additional arguments to the etcd container.
// They will be appended last to the command line.
func WithAdditionalArgs(args ...string) Option {
	return func(o *options) {
		o.AdditionalArgs = args
	}
}

// WithDataDir is an option to mount the data directory, which is located at /data.etcd.
func WithDataDir() Option {
	return func(o *options) {
		o.MountDataDir = true
	}
}

// WithNodes is an option to set the nodes of the etcd cluster.
// If only one node is specified, the cluster will be a single-node cluster.
func WithNodes(nodes ...string) Option {
	return func(o *options) {
		o.Nodes = nodes
	}
}

// withCurrentNode is an option to set the current node index.
// It's an internal option and should not be used by the user.
func withCurrentNode(i int) Option {
	return func(o *options) {
		o.currentNode = i
	}
}

// withClusterNetwork is an option to set the cluster network.
// It's an internal option and should not be used by the user.
func withClusterNetwork(n *testcontainers.DockerNetwork) Option {
	return func(o *options) {
		o.clusterNetwork = n
	}
}

// WithClusterToken is an option to set the cluster token.
func WithClusterToken(token string) Option {
	return func(o *options) {
		o.ClusterToken = token
	}
}

func withClusterOptions(opts []Option) Option {
	return func(o *options) {
		for _, opt := range opts {
			opt(o)
		}
	}
}
