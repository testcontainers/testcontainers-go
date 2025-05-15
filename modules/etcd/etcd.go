package etcd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
)

const (
	clientPort          = "2379"
	peerPort            = "2380"
	dataDir             = "/data.etcd"
	defaultClusterToken = "mys3cr3ttok3n"
	scheme              = "http"
)

// EtcdContainer represents the etcd container type used in the module. It can be used to create a single-node instance or a cluster.
// For the cluster, the first node creates the cluster and the other nodes join it as child nodes.
type EtcdContainer struct {
	testcontainers.Container
	// childNodes contains the child nodes of the current node, forming a cluster
	childNodes []*EtcdContainer
	opts       options
}

// Terminate terminates the etcd container, its child nodes, and the network in which the cluster is running
// to communicate between the nodes.
func (c *EtcdContainer) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	var errs []error

	// child nodes has no other children
	for i, child := range c.childNodes {
		if err := child.Terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate child node(%d): %w", i, err))
		}
	}

	if c.Container != nil {
		if err := c.Container.Terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate cluster node: %w", err))
		}
	}

	// remove the cluster network if it was created, but only for the first node
	// we could check if the current node is the first one (index 0),
	// and/or check that there are no child nodes
	if c.opts.clusterNetwork != nil && c.opts.currentNode == 0 {
		if err := c.opts.clusterNetwork.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("remove cluster network: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Run creates an instance of the etcd container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*EtcdContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{clientPort, peerPort},
		Cmd:          []string{},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	settings := defaultOptions(&req)
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&settings)
		}
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	clusterOpts, err := configureCluster(ctx, &settings, opts)
	if err != nil {
		return nil, fmt.Errorf("configure cluster: %w", err)
	}

	// configure CMD with the nodes
	genericContainerReq.Cmd = configureCMD(settings)

	// Initialise the etcd container with the current settings.
	// The cluster network, if needed, is already part of the settings,
	// so the following error handling returns a partially initialised container,
	// allowing the caller to clean up the resources with the Terminate method.
	c := &EtcdContainer{opts: settings}

	if settings.clusterNetwork != nil {
		// apply the network to the current node
		err := tcnetwork.WithNetwork([]string{settings.nodeNames[settings.currentNode]}, settings.clusterNetwork)(&genericContainerReq)
		if err != nil {
			return c, fmt.Errorf("with network: %w", err)
		}
	}

	if c.Container, err = testcontainers.GenericContainer(ctx, genericContainerReq); err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	// only the first node creates the cluster
	if settings.currentNode == 0 {
		for i := 1; i < len(settings.nodeNames); i++ {
			// move to the next node
			childNode, err := Run(ctx, req.Image, append(clusterOpts, withCurrentNode(i))...)
			if err != nil {
				// return the parent cluster node and the error, so the caller can clean up.
				return c, fmt.Errorf("run cluster node: %w", err)
			}

			c.childNodes = append(c.childNodes, childNode)
		}
	}

	return c, nil
}

// configureCluster configures the cluster settings, ensuring that the cluster is properly configured with the necessary network and options,
// avoiding duplicate application of options to be passed to the successive nodes.
func configureCluster(ctx context.Context, settings *options, opts []testcontainers.ContainerCustomizer) ([]testcontainers.ContainerCustomizer, error) {
	var clusterOpts []testcontainers.ContainerCustomizer
	if len(settings.nodeNames) == 0 {
		return clusterOpts, nil
	}

	// pass cluster options to each node
	etcdOpts := []Option{}
	for _, opt := range opts {
		// if the option is of type Option, it won't be applied to the settings
		// this prevents the same option from being applied multiple times (e.g. updating the current node)
		if apply, ok := opt.(Option); ok {
			etcdOpts = append(etcdOpts, apply)
		} else {
			clusterOpts = append(clusterOpts, opt)
		}
	}

	if settings.clusterNetwork == nil { // the first time the network is created
		newNetwork, err := tcnetwork.New(ctx)
		if err != nil {
			return clusterOpts, fmt.Errorf("new network: %w", err)
		}

		// set the network for the first node
		settings.clusterNetwork = newNetwork

		clusterOpts = append(clusterOpts, withClusterNetwork(newNetwork)) // save the network for the next nodes
	}

	// we finally need to re-apply all the etcd-specific options
	clusterOpts = append(clusterOpts, withClusterOptions(etcdOpts))

	return clusterOpts, nil
}

// configureCMD configures the etcd command line arguments, based on the settings provided,
// in order to create a cluster or a single-node instance.
func configureCMD(settings options) []string {
	cmds := []string{"etcd"}

	if len(settings.nodeNames) == 0 {
		cmds = append(cmds, "--name=default",
			"--listen-client-urls="+scheme+"://0.0.0.0:"+clientPort,
			"--advertise-client-urls="+scheme+"://0.0.0.0:"+clientPort,
		)
	} else {
		clusterCmds := []string{
			"--name=" + settings.nodeNames[settings.currentNode],
			"--initial-advertise-peer-urls=" + scheme + "://" + settings.nodeNames[settings.currentNode] + ":" + peerPort,
			"--advertise-client-urls=" + scheme + "://" + settings.nodeNames[settings.currentNode] + ":" + clientPort,
			"--listen-peer-urls=" + scheme + "://0.0.0.0:" + peerPort,
			"--listen-client-urls=" + scheme + "://0.0.0.0:" + clientPort,
			"--initial-cluster-state=new",
		}

		clusterStateValues := make([]string, len(settings.nodeNames))
		for i, node := range settings.nodeNames {
			clusterStateValues[i] = node + "=" + scheme + "://" + node + ":" + peerPort
		}
		clusterCmds = append(clusterCmds, "--initial-cluster="+strings.Join(clusterStateValues, ","))

		if settings.clusterToken != "" {
			clusterCmds = append(clusterCmds, "--initial-cluster-token="+settings.clusterToken)
		}

		cmds = append(cmds, clusterCmds...)
	}

	if settings.mountDataDir {
		cmds = append(cmds, "--data-dir="+dataDir)
	}

	cmds = append(cmds, settings.additionalArgs...)

	return cmds
}

// ClientEndpoint returns the client endpoint for the etcd container, and an error if any.
// For a cluster, it returns the client endpoint of the first node.
func (c *EtcdContainer) ClientEndpoint(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, clientPort)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

// ClientEndpoints returns the client endpoints for the etcd cluster.
func (c *EtcdContainer) ClientEndpoints(ctx context.Context) ([]string, error) {
	endpoint, err := c.ClientEndpoint(ctx)
	if err != nil {
		return nil, err
	}

	endpoints := []string{endpoint}

	for _, node := range c.childNodes {
		endpoint, err := node.ClientEndpoint(ctx)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// PeerEndpoint returns the peer endpoint for the etcd container, and an error if any.
// For a cluster, it returns the peer endpoint of the first node.
func (c *EtcdContainer) PeerEndpoint(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, peerPort)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

// PeerEndpoints returns the peer endpoints for the etcd cluster.
func (c *EtcdContainer) PeerEndpoints(ctx context.Context) ([]string, error) {
	endpoint, err := c.PeerEndpoint(ctx)
	if err != nil {
		return nil, err
	}

	endpoints := []string{endpoint}

	for _, node := range c.childNodes {
		endpoint, err := node.PeerEndpoint(ctx)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}
