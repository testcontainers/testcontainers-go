# Chroma

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

## Introduction

The Testcontainers module for Chroma.

## Resources

- [Chroma Docs](https://docs.trychroma.com/docs/overview/getting-started) - Chroma official documentation.
- [Chroma Cookbook](http://cookbook.chromadb.dev) - Community-driven Chroma cookbook.

## Adding this module to your project dependencies

Please run the following command to add the Chroma module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/chroma
```

## Usage example

<!--codeinclude-->
[Creating a Chroma container](../../modules/chroma/examples_test.go) inside_block:runChromaContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Chroma module exposes one entrypoint function to create the Chroma container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ChromaContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "chromadb/chroma:1.4.0")`.

### Container Options

When starting the Chroma container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Chroma container exposes the following methods:

#### REST Endpoint

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

This method returns the REST endpoint of the Chroma container, using the default `8000` port.

<!--codeinclude-->
[Get REST endpoint](../../modules/chroma/chroma_test.go) inside_block:restEndpoint
<!--/codeinclude-->

## Examples

### Getting a Chroma client

The following example demonstrates how to create a Chroma client using the Chroma module.

First of all, you need to import the Chroma module and the chroma-go client:

```golang
import (
    chromago "github.com/amikos-tech/chroma-go/pkg/api/v2"
)
```

Then, you can create a Chroma client using the Chroma module:

<!--codeinclude-->
[Get the client](../../modules/chroma/examples/examples_test.go) inside_block:getClient
<!--/codeinclude-->

### Working with Collections

The Chroma Go client's default embedding function needs the ONNX Runtime native library.
To run these examples locally, run `.github/scripts/modules/chroma/examples/install-dependencies.sh`
from the repository root and execute the `export CHROMAGO_ONNX_RUNTIME_PATH=...` command it prints.
The script requires Python 3 with pip and downloads ONNX Runtime 1.22.0 from PyPI.
CI runs this setup automatically before testing the examples.

<!--codeinclude-->
[Create Collection](../../modules/chroma/examples/examples_test.go) inside_block:createCollection
[List Collections](../../modules/chroma/examples/examples_test.go) inside_block:listCollections
[Add Data to Collection](../../modules/chroma/examples/examples_test.go) inside_block:addData
[Query Collection](../../modules/chroma/examples/examples_test.go) inside_block:queryCollection
[Delete Collection](../../modules/chroma/examples/examples_test.go) inside_block:deleteCollection
<!--/codeinclude-->
