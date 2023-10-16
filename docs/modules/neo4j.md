# Neo4j

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

## Introduction

The Testcontainers module for [Neo4j](https://neo4j.com/), the leading graph platform.

## Adding this module to your project dependencies

Please run the following command to add the Neo4j module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/neo4j
```

## Usage example

Running Neo4j as a single-instance server, with the [APOC plugin](https://neo4j.com/developer/neo4j-apoc/) enabled:

<!--codeinclude-->
[Creating a Neo4j container](../../modules/neo4j/examples_test.go) inside_block:runNeo4jContainer
<!--/codeinclude-->

## Module Reference

The Neo4j module exposes one entrypoint function to create the Neo4j container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Neo4jContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Ports

These are the ports used by the Neo4j container:

<!--codeinclude-->
[Container Ports](../../modules/neo4j/neo4j.go) inside_block:containerPorts
<!--/codeinclude-->

### Container Options

When starting the Neo4j container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Neo4j Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Neo4j. E.g. `testcontainers.WithImage("docker.io/neo4j:4.4")`.

By default, the container will use the following Docker image:

<!--codeinclude-->
[Default Docker image](../../modules/neo4j/neo4j.go) inside_block:defaultImage
<!--/codeinclude-->

{% include "../features/common_functional_options.md" %}

#### Logger

This option sets a custom logger to be used by the container. Consider calling this before other `With` functions as these may generate logs.

!!!info
    The logger must implement the `testcontainers.Logger` interface.

<!--codeinclude-->
[Including a custom logger](../../modules/neo4j/neo4j_test.go) inside_block:withSettings
<!--/codeinclude-->

#### Authentication

By default, the Neo4j container will be started with authentication disabled. If you need to enable authentication, you can
use the `WithAuthentication(pwd string)` option.

By default, the container will not use authentication, automatically prepending the `WithoutAuthentication` option to the options list.

#### Plugins

By default, the Neo4j container will start without any Labs plugins enabled, but you can enable them using the `WithLabsPlugin` optional function.

<!--codeinclude-->
[Adding Labs Plugins](../../modules/neo4j/neo4j_test.go) inside_block:withLabsPlugin
<!--/codeinclude-->

The list of available plugins is:

<!--codeinclude-->
[Labs plugins](../../modules/neo4j/config.go) inside_block:labsPlugins
<!--/codeinclude-->

#### Settings

It's possible to add Neo4j a single configuration setting to the container.
The setting can be added as in the official Neo4j configuration, the function automatically translates the setting
name (e.g. ``dbms.tx_log.rotation.size`) into the format required by the Neo4j container.
This function can be called multiple times. A warning is emitted if a key is overwritten.

To pass multiple settings at once, the `WithNeo4jSettings` function is provided.

<!--codeinclude-->
[Adding settings](../../modules/neo4j/neo4j_test.go) inside_block:withSettings
<!--/codeinclude-->

!!!warning
    Credentials must be configured with the `WithAdminPassword` optional function.

### Container Methods

#### Bolt URL

The `BoltURL` method returns the connection string to connect to the Neo4j container instance using the Bolt port.
It returns a string with the format `neo4j://<host>:<port>`.

<!--codeinclude-->
[Connect to Neo4j](../../modules/neo4j/neo4j_test.go) inside_block:boltURL
<!--/codeinclude-->
