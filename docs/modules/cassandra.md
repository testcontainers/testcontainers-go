# Cassandra

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.26.0"><span class="tc-version">:material-tag: v0.26.0</span></a>

## Introduction

The Testcontainers module for Cassandra.

## Adding this module to your project dependencies

Please run the following command to add the Cassandra module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/cassandra
```

## Usage example

<!--codeinclude-->
[Creating a Cassandra container](../../modules/cassandra/examples_test.go) inside_block:runCassandraContainer
<!--/codeinclude-->

## Module reference

The Cassandra module exposes one entrypoint function to create the Cassandra container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CassandraContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Cassandra container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Cassandra Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Cassandra. E.g. `testcontainers.WithImage("cassandra:4.1.3")`.

{% include "../features/common_functional_options.md" %}

#### Init Scripts

If you would like to do additional initialization in the Cassandra container, add one or more `*.cql` or `*.sh` scripts to the container request with the `WithInitScripts` function.
Those files will be copied after the container is created but before it's started under root directory.

An example of a `*.sh` script that creates a keyspace and table is shown below:

<!--codeinclude-->
[Init script content](../../modules/cassandra/testdata/init.sh)
<!--/codeinclude-->

#### Database configuration

In the case you have a custom config file for Cassandra, it's possible to copy that file into the container before it's started, using the `WithConfigFile(cfgPath string)` function.

!!!warning
    You should provide a valid Cassandra configuration file, otherwise the container will fail to start.

### Container Methods

The Cassandra container exposes the following methods:

#### ConnectionHost

This method returns the host and port of the Cassandra container, using the default, `9042/tcp` port. E.g. `localhost:9042`

<!--codeinclude-->
[Get connection host](../../modules/cassandra/cassandra_test.go) inside_block:connectionHost
<!--/codeinclude-->
