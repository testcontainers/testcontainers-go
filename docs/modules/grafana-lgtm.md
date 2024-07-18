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
[Creating a Grafana LGTM container](../../modules/grafana-lgtm/examples_test.go) inside_block:runGrafanaLGTMContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Grafana LGTM module exposes one entrypoint function to create the Grafana LGTM container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GrafanaLGTMContainer, error)
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

The Grafana LGTM container exposes the following methods:

#### GrafanaURL

This method returns the HTTP URL to connect to Grafana, using the default `3000` port.

#### LokiURL

This method returns the HTTP URL to connect to Loki, using the default `3100` port.

#### TempoURL

This method returns the HTTP URL to connect to Tempo, using the default `3200` port.

#### Otel HTTP URL

This method returns the URL to connect to Otel using HTTP, using the default `4318` port.

#### Otel gRPC URL

This method returns the URL to connect to Otel using gRPC, using the default `4317` port.

#### Prometheus URL

This method returns the URL to connect to Prometheus, using the default `9090` port.
