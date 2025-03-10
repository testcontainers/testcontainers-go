# etcd

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

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

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

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

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "bitnami/etcd:latest")`.

{% include "../features/common_functional_options.md" %}

#### WithAdditionalArgs

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

You can pass additional arguments to the etcd container by using the `WithAdditionalArgs` option. The arguments are passed to the CMD of the etcd container.

#### WithDataDir

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

You can set the data directory for the etcd container by using the `WithDataDir` boolean option. The data directory where the etcd data is stored is `/data.etcd`.

#### WithNodes

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

You can set the number of nodes for the etcd cluster by using the `WithNodes` option, passing the node names for each of the nodes. Single-node clusters are not allowed,
for that reason the functional option receives three string arguments: the first node, the second node, and a variadic argument for the rest of the nodes.
The module starts a container for each node, having the first node a reference to the other nodes. E.g. `WithNodes("etcd-1", "etcd-2")`, `WithNodes("etcd-1", "etcd-2", "etcd-3")` and so on.

The module creates a Docker network for the etcd cluster, and the nodes are connected to this network, so that they can communicate with each other through the network.

#### WithClusterToken

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

Sets the cluster token for the etcd cluster. The cluster token is used to identify the etcd cluster. The default value is `mys3cr3ttok3n`.
The etcd container holds a reference to the cluster token, so you can use it with e.g. `ctr.ClusterToken`.

### Container Methods

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

The etcd container exposes the following methods:

#### ClientEndpoint

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

Returns the client endpoint for the etcd container and an error, if any. In the case of a cluster, it returns the client endpoint for the first node.

#### PeerEndpoint

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

Returns the peer endpoint for the etcd container and an error, if any. In the case of a cluster, it returns the peer endpoint for the first node.
