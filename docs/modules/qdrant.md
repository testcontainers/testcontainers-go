# Qdrant

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

## Introduction

The Testcontainers module for Qdrant.

## Adding this module to your project dependencies

Please run the following command to add the Qdrant module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/qdrant
```

## Usage example

<!--codeinclude-->
[Creating a Qdrant container](../../modules/qdrant/examples_test.go) inside_block:runQdrantContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Qdrant module exposes one entrypoint function to create the Qdrant container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*QdrantContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Qdrant container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "qdrant/qdrant:v1.7.4")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Qdrant container exposes the following methods:

#### REST Endpoint

This method returns the REST endpoint of the Qdrant container, using the default `6333` port.

<!--codeinclude-->
[Get REST endpoint](../../modules/qdrant/qdrant_test.go) inside_block:restEndpoint
<!--/codeinclude-->

#### Web UI Endpoint

This method returns the Web UI endpoint of the Qdrant container (`/dashboard`), using the default `6333` port.

<!--codeinclude-->
[Get Web UI endpoint](../../modules/qdrant/qdrant_test.go) inside_block:webUIEndpoint
<!--/codeinclude-->

#### gRPC Endpoint

This method returns the gRPC endpoint of the Qdrant container, using the default `6334` port.

<!--codeinclude-->
[Get gRPC endpoint](../../modules/qdrant/qdrant_test.go) inside_block:gRPCEndpoint
<!--/codeinclude-->

### Full Example

Here you can find a full example on how to use the qdrant-go module to perform operations with Qdrant, as seen in the [examples provided by the module](https://github.com/qdrant/go-client/blob/76db566382ed656a920fa273db1a58eec2417dcd/examples/main.go#L1) itself:

<!--codeinclude-->
[Full Example](../../modules/qdrant/examples_test.go) inside_block:fullExample
<!--/codeinclude-->