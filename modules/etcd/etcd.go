package etcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
)

const (
	ClientPort          = "2379"
	PeerPort            = "2380"
	DataDir             = "/data.etcd"
	DefaultClusterToken = "mys3cr3ttok3n"
)

// EtcdContainer represents the etcd container type used in the module
type EtcdContainer struct {
	testcontainers.Container
	ClusterToken string
	Nodes        []*EtcdContainer
}

// Run creates an instance of the etcd container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*EtcdContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
		Cmd:   []string{},
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

	if settings.clusterNetwork != nil {
		// apply the network to the current node
		err := tcnetwork.WithNetwork([]string{settings.Nodes[settings.currentNode]}, settings.clusterNetwork)(&genericContainerReq)
		if err != nil {
			return nil, fmt.Errorf("with network: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *EtcdContainer
	if container != nil {
		c = &EtcdContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	// only the first node creates the cluster
	if settings.currentNode == 0 {
		for i := 1; i < len(settings.Nodes); i++ {
			// move to the next node
			clusterNode, err := Run(ctx, req.Image, append(clusterOpts, withCurrentNode(i))...)
			if err != nil {
				// return the parent cluster node and the error, so the caller can clean up.
				return c, fmt.Errorf("run cluster node: %w", err)
			}

			c.Nodes = append(c.Nodes, clusterNode)
		}
	}

	c.ClusterToken = settings.ClusterToken

	return c, nil
}

// configureCluster configures the cluster settings, ensuring that the cluster is properly configured with the necessary network and options,
// avoiding duplicate application of options to be passed to the successive nodes.
func configureCluster(ctx context.Context, settings *options, opts []testcontainers.ContainerCustomizer) ([]testcontainers.ContainerCustomizer, error) {
	var clusterOpts []testcontainers.ContainerCustomizer
	if len(settings.Nodes) <= 1 {
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

	scheme := "http"

	if len(settings.Nodes) < 1 {
		cmds = append(cmds, "--name=default")
	} else {
		clusterCmds := []string{
			"--name=" + settings.Nodes[settings.currentNode],
			"--initial-advertise-peer-urls=" + scheme + "://" + settings.Nodes[settings.currentNode] + ":" + PeerPort,
			"--advertise-client-urls=" + scheme + "://" + settings.Nodes[settings.currentNode] + ":" + ClientPort,
			"--listen-peer-urls=" + scheme + "://0.0.0.0:" + PeerPort,
			"--listen-client-urls=" + scheme + "://0.0.0.0:" + ClientPort,
			"--initial-cluster-state=new",
		}

		clusterStateValues := []string{}
		for _, node := range settings.Nodes {
			clusterStateValues = append(clusterStateValues, node+"="+scheme+"://"+node+":"+PeerPort)
		}
		clusterCmds = append(clusterCmds, "--initial-cluster="+strings.Join(clusterStateValues, ","))

		if settings.ClusterToken != "" {
			clusterCmds = append(clusterCmds, "--initial-cluster-token="+settings.ClusterToken)
		}

		cmds = append(cmds, clusterCmds...)
	}

	if settings.mountDataDir {
		cmds = append(cmds, "--data-dir="+DataDir)
	}

	cmds = append(cmds, settings.AdditionalArgs...)

	return cmds
}

// ClientEndpoint returns the client endpoint for the etcd container, and an error if any.
// For a cluster, it returns the client endpoint of the first node.
func (c *EtcdContainer) ClientEndpoint(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, ClientPort)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

// MustClientEndpoint returns the client endpoint for the etcd container.
// For a cluster, it returns the client endpoint of the first node.
// This method panics if an error occurs.
func (c *EtcdContainer) MustClientEndpoint(ctx context.Context) string {
	endpoint, err := c.ClientEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return endpoint
}

// MustClientEndpoints returns the client endpoints for the etcd cluster.
// This method panics if an error occurs.
func (c *EtcdContainer) MustClientEndpoints(ctx context.Context) []string {
	endpoints := []string{c.MustClientEndpoint(ctx)}

	for _, node := range c.Nodes {
		endpoints = append(endpoints, node.MustClientEndpoint(ctx))
	}

	return endpoints
}

// PeerEndpoint returns the peer endpoint for the etcd container, and an error if any.
// For a cluster, it returns the peer endpoint of the first node.
func (c *EtcdContainer) PeerEndpoint(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, PeerPort)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

// MustPeerEndpoint returns the peer endpoint for the etcd container.
// For a cluster, it returns the peer endpoint of the first node.
// This method panics if an error occurs.
func (c *EtcdContainer) MustPeerEndpoint(ctx context.Context) string {
	endpoint, err := c.PeerEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	return endpoint
}

// MustPeerEndpoints returns the peer endpoints for the etcd cluster.
// This method panics if an error occurs.
func (c *EtcdContainer) MustPeerEndpoints(ctx context.Context) []string {
	endpoints := []string{c.MustPeerEndpoint(ctx)}

	for _, node := range c.Nodes {
		endpoints = append(endpoints, node.MustPeerEndpoint(ctx))
	}

	return endpoints
}
