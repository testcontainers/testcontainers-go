# Milvus

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

## Introduction

The Testcontainers module for Milvus.

## Adding this module to your project dependencies

Please run the following command to add the Milvus module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/milvus
```

## Usage example

<!--codeinclude-->
[Creating a Milvus container](../../modules/milvus/examples_test.go) inside_block:runMilvusContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Milvus module exposes one entrypoint function to create the Milvus container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MilvusContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Milvus container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "milvusdb/milvus:v2.3.9")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Milvus container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the Milvus container, using the default `19530` port.

<!--codeinclude-->
[Get connection string](../../modules/milvus/milvus_test.go) inside_block:connectionString
<!--/codeinclude-->

## Examples

### Creating collections

This example shows the usage of the Milvus module to create and retrieve collections.

<!--codeinclude-->
[Create collections](../../modules/milvus/examples_test.go) inside_block:createCollections
<!--/codeinclude-->
