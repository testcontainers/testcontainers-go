# MockServer

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

## Introduction

The Testcontainers module for MockServer. MockServer can be used to mock HTTP services by matching requests against user-defined expectations.

## Adding this module to your project dependencies

Please run the following command to add the MockServer module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mockserver
```

## Usage example

<!--codeinclude-->
[Creating a MockServer container](../../modules/mockserver/examples_test.go) inside_block:runMockServerContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The MockServer module exposes one entrypoint function to create the MockServer container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MockServerContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the MockServer container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "mockserver/mockserver:5.15.0")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The MockServer container exposes the following methods:

#### URL

The `URL` method returns the url string to connect to the MockServer container.
It returns a string with the format `http://<host>:<port>`.

It can be used to configure a MockServer client (`github.com/BraspagDevelopers/mock-server-client`), e.g.:

<!--codeinclude-->
[Using URL with the MockServer client](../../modules/mockserver/examples_test.go) inside_block:connectToMockServer
<!--/codeinclude-->
