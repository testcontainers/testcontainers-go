# Chroma

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Chroma.

## Adding this module to your project dependencies

Please run the following command to add the Chroma module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/chroma
```

## Usage example

<!--codeinclude-->
[Creating a Chroma container](../../modules/chroma/examples_test.go) inside_block:runChromaContainer
<!--/codeinclude-->

## Module reference

The Chroma module exposes one entrypoint function to create the Chroma container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ChromaContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Chroma container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Chroma Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Chroma. E.g. `testcontainers.WithImage("chromadb/chroma:0.4.22.dev44")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Chroma container exposes the following methods:

#### REST Endpoint

This method returns the REST endpoint of the Chroma container, using the default `8000` port.

<!--codeinclude-->
[Get REST endpoint](../../modules/chroma/chroma_test.go) inside_block:restEndpoint
<!--/codeinclude-->

## Examples

### Getting a Chroma client

The following example demonstrates how to create a Chroma client using the Chroma module.

First of all, you need to import the Chroma module and the Swagger client:

```golang
import (
    chromago "github.com/amikos-tech/chroma-go"
	chromaopenapi "github.com/amikos-tech/chroma-go/swagger"
)
```

Then, you can create a Chroma client using the Chroma module:

<!--codeinclude-->
[Get the client](../../modules/chroma/examples_test.go) inside_block:createClient
<!--/codeinclude-->

### Working with Collections

<!--codeinclude-->
[Create Collection](../../modules/chroma/examples_test.go) inside_block:createCollection
[List Collections](../../modules/chroma/examples_test.go) inside_block:listCollections
[Delete Collection](../../modules/chroma/examples_test.go) inside_block:deleteCollection
<!--/codeinclude-->