# Timeplus

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Timeplus.

[Timeplus](https://www.timeplus.com/) is a simple, powerful, and cost-efficient stream processing platform. It is compatible with the ClickHouse wire protocol, so any ClickHouse client library can connect to a Timeplus instance.

## Adding this module to your project dependencies

Please run the following command to add the Timeplus module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/timeplus
```

## Usage example

<!--codeinclude-->
[Creating a Timeplus container](../../modules/timeplus/examples_test.go) inside_block:runTimeplusContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Timeplus module exposes one entrypoint function to create the Timeplus container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "timeplus/timeplusd:2.3")`.

### Container Ports

Here you can find the list with the default ports used by the Timeplus container.

<!--codeinclude-->
[Container Ports](../../modules/timeplus/timeplus.go) inside_block:containerPorts
<!--/codeinclude-->

### Container Options

When starting the Timeplus container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Timeplus container exposes the following methods:

#### HTTPEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the HTTP endpoint of the Timeplus container for the ClickHouse-compatible HTTP API (port 8123), in the form `http://host:port`.

<!--codeinclude-->
[Get HTTP endpoint](../../modules/timeplus/examples_test.go) inside_block:ExampleContainer_HTTPEndpoint
<!--/codeinclude-->

!!!info
    Because Timeplus is wire-protocol compatible with ClickHouse, you can use any ClickHouse client library to connect to a Timeplus container via the HTTP endpoint returned by `HTTPEndpoint`.
