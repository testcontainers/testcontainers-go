# Grafana LGTM

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.33.0"><span class="tc-version">:material-tag: v0.33.0</span></a>

## Introduction

The Testcontainers module for Grafana LGTM.

## Adding this module to your project dependencies

Please run the following command to add the Grafana module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/grafana-lgtm
```

## Usage example

<!--codeinclude-->
[Creating a Grafana LGTM container](../../modules/grafana-lgtm/examples_test.go) inside_block:runGrafanaLGTMContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.33.0"><span class="tc-version">:material-tag: v0.33.0</span></a>

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

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "grafana/otel-lgtm:0.6.0")`.

#### Admin Credentials

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.33.0"><span class="tc-version">:material-tag: v0.33.0</span></a>

If you need to set different admin credentials in the Grafana LGTM container, you can set them using the `WithAdminCredentials(user, password)` option.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Grafana LGTM container exposes the following methods:

!!!info
    All the endpoint methods return their endpoints in the format `<host>:<port>`, so please use them accordingly to configure your client.

#### Grafana Endpoint

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.33.0"><span class="tc-version">:material-tag: v0.33.0</span></a>

The `HttpEndpoint(ctx)` method returns the HTTP endpoint to connect to Grafana, using the default `3000` port. The same method with the `Must` prefix returns just the endpoint, and panics if an error occurs.

#### Loki Endpoint

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.33.0"><span class="tc-version">:material-tag: v0.33.0</span></a>

The `LokiEndpoint(ctx)` method returns the HTTP endpoint to connect to Loki, using the default `3100` port. The same method with the `Must` prefix returns just the endpoint, and panics if an error occurs.

#### Tempo Endpoint

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.33.0"><span class="tc-version">:material-tag: v0.33.0</span></a>

The `TempoEndpoint(ctx)` method returns the HTTP endpoint to connect to Tempo, using the default `3200` port. The same method with the `Must` prefix returns just the endpoint, and panics if an error occurs.

#### Otel HTTP Endpoint

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.33.0"><span class="tc-version">:material-tag: v0.33.0</span></a>

The `OtelHTTPEndpoint(ctx)` method returns the endpoint to connect to Otel using HTTP, using the default `4318` port. The same method with the `Must` prefix returns just the endpoint, and panics if an error occurs.

#### Otel gRPC Endpoint

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.33.0"><span class="tc-version">:material-tag: v0.33.0</span></a>

The `OtelGRPCEndpoint(ctx)` method returns the endpoint to connect to Otel using gRPC, using the default `4317` port. The same method with the `Must` prefix returns just the endpoint, and panics if an error occurs.

#### Prometheus Endpoint

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.33.0"><span class="tc-version">:material-tag: v0.33.0</span></a>

The `PrometheusHttpEndpoint(ctx)` method returns the endpoint to connect to Prometheus, using the default `9090` port. The same method with the `Must` prefix returns just the endpoint, and panics if an error occurs.

## Examples

### Traces, Logs and Prometheus metrics for a simple Go process

In this example, a simple application is created to generate traces, logs, and Prometheus metrics.
The application sends data to Grafana LGTM, and the Otel SDK is used to send the data.
The example demonstrates how to set up the Otel SDK and run the Grafana LGTM module,
configuring the Otel library to send data to Grafana LGTM thanks to the endpoints provided by the Grafana LGTM container.

<!--codeinclude-->
[App sending Otel data](../../modules/grafana-lgtm/examples_test.go) inside_block:rollDiceApp
[Setup Otel SDK](../../modules/grafana-lgtm/examples_test.go) inside_block:setupOTelSDK
[Run the Grafana LGTM container](../../modules/grafana-lgtm/examples_test.go) inside_block:ExampleRun_otelCollector
<!--/codeinclude-->