# ScyllaDB

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

## Introduction

The Testcontainers module for ScyllaDB, a NoSQL database fully compatible with Apache Cassandra and DynamoDB, allows you
to create a ScyllaDB container for testing purposes.

## Adding this module to your project dependencies

Please run the following command to add the ScyllaDB module to your Go dependencies:

```shell
go get github.com/testcontainers/testcontainers-go/modules/scylladb
```

## Usage example

<!--codeinclude-->
[Creating a ScyllaDB container](../../modules/scylladb/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

The ScyllaDB module exposes one entrypoint function to create the ScyllaDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ScyllaDBContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

!!! info
    By default, we add the `--developer-mode=1` flag to the ScyllaDB container to disable the various checks Scylla
    performs.
    Also in scenarios in which static partitioning is not desired - like mostly-idle cluster without hard latency
    requirements, the `--overprovisioned` command-line option is recommended. This enables certain optimizations for ScyllaDB
    to run efficiently in an overprovisioned environment. You can change it by using the `WithCustomCommand` function.

### Container Options

When starting the ScyllaDB container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example:

```golang
scylladb.Run(context.Background(), "scylladb/scylla:6.2.1")
// OR
scylladb.Run(context.Background(), "scylladb/scylla:5.6")
```

{% include "../features/common_functional_options.md" %}

#### With Database Configuration File (scylla.yaml)

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

In the case you have a custom config file for ScyllaDB, it's possible to copy that file into the container before it's
started, using the `WithConfig(r io.Reader)` function.

<!--codeinclude-->
[With Config YAML](../../modules/scylladb/examples_test.go) inside_block:runScyllaDBContainerWithConfig
<!--/codeinclude-->
!!!warning
    You should provide a valid ScyllaDB configuration file as an `io.Reader` when using the function, otherwise the container will fail to
    start. The configuration file should be a valid YAML file and follows
    the [ScyllaDB configuration file](https://github.com/scylladb/scylladb/blob/master/conf/scylla.yaml).

#### With Shard Awareness

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

If you want to test ScyllaDB with shard awareness, you can use the `WithShardAwareness` function. This function will
configure the ScyllaDB container to use the `19042` port and ask the container to wait until the port is ready.

<!--codeinclude-->
[With Shard Awareness](../../modules/scylladb/examples_test.go) inside_block:runScyllaDBContainerWithShardAwareness
<!--/codeinclude-->

#### With Alternator (DynamoDB Compatible API)

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

If you want to test ScyllaDB with the Alternator API, you can use the `WithAlternator` function. This function will
configure the ScyllaDB container to use the port any port you want and ask the container to wait until the port is
ready.
By default, you can choose the port `8000`.

<!--codeinclude-->
[With Alternator API](../../modules/scylladb/examples_test.go) inside_block:runScyllaDBContainerWithAlternator
<!--/codeinclude-->

#### With Custom Commands

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

If you need to pass any flag to the ScyllaDB container, you can use the `WithCustomCommand` function. This also rewrites
predefined commands like `--developer-mode=1`. You can check
the [ScyllaDB Docker Best Practices](https://opensource.docs.scylladb.com/stable/operating-scylla/procedures/tips/best-practices-scylla-on-docker.html) for more information.

<!--codeinclude-->
[With Custom Commands](../../modules/scylladb/examples_test.go) inside_block:runScyllaDBContainerWithCustomCommands
<!--/codeinclude-->

### Container Methods

The ScyllaDB container exposes the following methods:

#### ConnectionHost methods

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

There exist three methods to get the host and port of the ScyllaDB container, depending on the feature you want.

If you just want to test it with a single node and a single core, you can use the `NonShardAwareConnectionHost` method. However, if you're planning
to use more than one core, you should use the `ShardAwareConnectionHost` method, which uses the **shard-awareness** `19042` port.

Else, if you're planning to use the **Alternator** API, you should use the `AlternatorConnectionHost` method,
which uses the default port `8000`.

<!--codeinclude-->
[Non-shard-aware connection host](../../modules/scylladb/examples_test.go) inside_block:scyllaDbNonShardAwareConnectionHost
[Shard-aware connection host](../../modules/scylladb/examples_test.go) inside_block:scyllaDbShardAwareConnectionHost
[Alternator host](../../modules/scylladb/examples_test.go) inside_block:scyllaDbAlternatorConnectionHost
<!--/codeinclude-->
