# etcd

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for etcd.

## Adding this module to your project dependencies

Please run the following command to add the etcd module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/etcd
```

## Usage example

<!--codeinclude-->
[Creating a etcd container](../../modules/etcd/examples_test.go) inside_block:runetcdContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The etcd module exposes one entrypoint function to create the etcd container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*etcdContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the etcd container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different etcd Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "bitnami/etcd:latest")`.

{% include "../features/common_functional_options.md" %}

#### WithAdditionalArgs

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

You can pass additional arguments to the etcd container by using the `WithAdditionalArgs` option. The arguments are passed to the CMD of the etcd container.

#### WithDataDir

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

You can set the data directory for the etcd container by using the `WithDataDir` boolean option. The data directory where the etcd data is stored is `/data.etcd`.

#### WithNodes

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

You can set the number of nodes for the etcd cluster by using the `WithNodes` option, passing the node names for each of the nodes. The module starts a container for each node,
having the first node a reference to the other nodes. E.g. `WithNodes('etcd-1', 'etcd-2)`.

The module creates a Docker network for the etcd cluster, and the nodes are connected to this network, so that they can communicate with each other through the network.

#### WithClusterToken

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the cluster token for the etcd cluster. The cluster token is used to identify the etcd cluster. The default value is `mys3cr3ttok3n`.
The etcd container holds a reference to the cluster token, so you can use it with e.g. `ctr.ClusterToken`.

### Container Methods

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The etcd container exposes the following methods:

#### ClientEndpoint

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the client endpoint for the etcd container and an error, if any. In the case of a cluster, it returns the client endpoint for the first node.

#### MustClientEndpoint

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the client endpoint for the etcd container. In the case of a cluster, it returns the client endpoint for the first node. This method panics if an error occurs.

#### MustClientEndpoints

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the client endpoints for all the nodes in the etcd cluster. In the case of single node, it returns the client endpoint for the first node. This method panics if an error occurs.

#### PeerEndpoint

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the peer endpoint for the etcd container and an error, if any. In the case of a cluster, it returns the peer endpoint for the first node.

#### MustPeerEndpoint

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the peer endpoint for the etcd container. In the case of a cluster, it returns the peer endpoint for the first node. This method panics if an error occurs.

#### MustPeerEndpoints

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the peer endpoints for all the nodes in the etcd cluster. In the case of single node, it returns the peer endpoint for the first node. This method panics if an error occurs.