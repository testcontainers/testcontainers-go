# MockServer

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

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

## Module reference

The MockServer module exposes one entrypoint function to create the MockServer container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MockServerContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the MockServer container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different MockServer Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for MockServer. E.g. `testcontainers.WithImage("mockserver/mockserver:5.15.0")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The MockServer container exposes the following methods:

#### URL

The `URL` method returns the url string to connect to the MockServer container.
It returns a string with the format `http://<host>:<port>`.

It can be use to configure a MockServer client (`github.com/BraspagDevelopers/mock-server-client`), e.g.:

<!--codeinclude-->
[Using URL with the MockServer client](../../modules/mockserver/examples_test.go) inside_block:connectToMockServer
<!--/codeinclude-->
