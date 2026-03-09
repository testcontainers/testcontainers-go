package etcd

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	currentNode    int
	clusterNetwork *testcontainers.DockerNetwork
	nodeNames      []string
	clusterToken   string
	additionalArgs []string
	mountDataDir   bool // flag needed to avoid extra calculations with the lifecycle hooks
}

func defaultOptions() options {
	return options{
		clusterToken: defaultClusterToken,
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Etcd container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithAdditionalArgs is an option to pass additional arguments to the etcd container.
// They will be appended last to the command line.
func WithAdditionalArgs(args ...string) Option {
	return func(o *options) error {
		o.additionalArgs = args
		return nil
	}
}

// WithDataDir is an option to mount the data directory, which is located at /data.etcd.
// The option will add a lifecycle hook to the container to change the permissions of the data directory.
func WithDataDir() Option {
	return func(o *options) error {
		// Avoid extra calculations with the lifecycle hooks
		o.mountDataDir = true
		return nil
	}
}

// WithNodes is an option to set the nodes of the etcd cluster.
// It should be used to create a cluster with more than one node.
func WithNodes(node1 string, node2 string, nodes ...string) Option {
	return func(o *options) error {
		o.nodeNames = append([]string{node1, node2}, nodes...)
		return nil
	}
}

// withCurrentNode is an option to set the current node index.
// It's an internal option and should not be used by the user.
func withCurrentNode(i int) Option {
	return func(o *options) error {
		o.currentNode = i
		return nil
	}
}

// withClusterNetwork is an option to set the cluster network.
// It's an internal option and should not be used by the user.
func withClusterNetwork(n *testcontainers.DockerNetwork) Option {
	return func(o *options) error {
		o.clusterNetwork = n
		return nil
	}
}

// WithClusterToken is an option to set the cluster token.
func WithClusterToken(token string) Option {
	return func(o *options) error {
		o.clusterToken = token
		return nil
	}
}

func withClusterOptions(opts []Option) Option {
	return func(o *options) error {
		for _, opt := range opts {
			if err := opt(o); err != nil {
				return fmt.Errorf("with cluster options: %w", err)
			}
		}
		return nil
	}
}
