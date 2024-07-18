# Grafana LGTM

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Grafana LGTM.

## Adding this module to your project dependencies

Please run the following command to add the Grafana module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/grafanalgtm
```

## Usage example

<!--codeinclude-->
[Creating a Grafana container](../../modules/grafana-lgtm/examples_test.go) inside_block:runGrafanaContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Grafana LGTM module exposes one entrypoint function to create the Grafana container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GrafanaContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Grafana LGTM container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Grafana LGTM Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "grafana/otel-lgtm:0.6.0")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Grafana container exposes the following methods:
