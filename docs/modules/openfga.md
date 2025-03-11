# OpenFGA

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.30.0"><span class="tc-version">:material-tag: v0.30.0</span></a>

## Introduction

The Testcontainers module for OpenFGA.

## Adding this module to your project dependencies

Please run the following command to add the OpenFGA module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/openfga
```

## Usage example

<!--codeinclude-->
[Creating a OpenFGA container](../../modules/openfga/examples_test.go) inside_block:runOpenFGAContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The OpenFGA module exposes one entrypoint function to create the OpenFGA container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OpenFGAContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the OpenFGA container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "openfga/openfga:v1.5.0")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The OpenFGA container exposes the following methods:

#### HttpEndpoint

This method returns the HTTP endpoint to connect to the OpenFGA container, using the `8080` port.

<!--codeinclude-->
[Get HTTP endpoint](../../modules/openfga/examples_test.go) inside_block:httpEndpoint
<!--/codeinclude-->

#### GrpcEndpoint

This method returns the gRPC endpoint to connect to the OpenFGA container, using the `8081` port.

#### Playground URL

In case you want to interact with the openfga playground, please use the `PlaygroundEndpoint` method, using the `3000` port.

<!--codeinclude-->
[Get Playground endpoint](../../modules/openfga/examples_test.go) inside_block:playgroundEndpoint
<!--/codeinclude-->

## Examples

### Writing an OpenFGA model

The following example shows how to write an OpenFGA model using the OpenFGA container.

<!--codeinclude-->
[Get Playground endpoint](../../modules/openfga/examples_test.go) inside_block:openFGAwriteModel
<!--/codeinclude-->
