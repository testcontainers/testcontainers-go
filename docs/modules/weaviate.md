# Weaviate

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

## Introduction

The Testcontainers module for Weaviate.

## Adding this module to your project dependencies

Please run the following command to add the Weaviate module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/weaviate
```

## Usage example

<!--codeinclude-->
[Creating a Weaviate container](../../modules/weaviate/examples_test.go) inside_block:runWeaviateContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Weaviate module exposes one entrypoint function to create the Weaviate container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*WeaviateContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Weaviate container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "semitechnologies/weaviate:1.23.9")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Weaviate container exposes the following methods:

#### HTTP Host Address

This method returns the Schema and Host for the Weaviate container, using the default `8080` port.

!!!info
    At the moment, the Weaviate module only supports the HTTP schema.

<!--codeinclude-->
[HTTP Host Address](../../modules/weaviate/weaviate_test.go) inside_block:httpHostAddress
<!--/codeinclude-->

## Examples

### Getting a Weaviate client

The following example demonstrates how to create a Weaviate client using the Weaviate module.

First of all, you need to import the Weaviate client:

```golang
import (
    "github.com/weaviate/weaviate-go-client/v4/weaviate"
)
```

Then, you can create a Weaviate client using the Weaviate module:

<!--codeinclude-->
[Get the client](../../modules/weaviate/examples_test.go) inside_block:createClientNoModules
[Get the client And Modules](../../modules/weaviate/examples_test.go) inside_block:createClientAndModules
<!--/codeinclude-->
