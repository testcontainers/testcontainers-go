# Pinecone

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Pinecone.

## Adding this module to your project dependencies

Please run the following command to add the Pinecone module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/pinecone
```

## Usage example

<!--codeinclude-->
[Creating a Pinecone container](../../modules/pinecone/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Pinecone module exposes one entrypoint function to create the Pinecone container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*PineconeContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Pinecone container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "ghcr.io/pinecone-io/pinecone-local:latest")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Pinecone container exposes the following methods:

#### HttpEndpoint

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `HttpEndpoint` method returns the location where the Pinecone container is listening.
It returns a string with the format `http://<host>:<port>`.

<!--codeinclude-->
[Connecting using HTTP](../../modules/pinecone/examples_test.go) inside_block:httpConnection
<!--/codeinclude-->
