# Presto

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [Presto](https://prestodb.io/), a distributed SQL query engine for big data.
Presto is designed to run interactive analytic queries against data sources of all sizes, from gigabytes to petabytes.

## Adding this module to your project dependencies

Please run the following command to add the Presto module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/presto
```

## Usage example

<!--codeinclude-->
[Creating a Presto container](../../modules/presto/examples_test.go) inside_block:runPrestoContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Presto module exposes one entrypoint function to create the Presto container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "prestodb/presto:0.286")`.

### Wait Strategy

The module waits for the Presto coordinator to finish starting up by polling
the `/v1/info` HTTP endpoint on port `8080/tcp` until the `starting` field in
the JSON response is `false`. A startup timeout of 2 minutes is applied.

### Container Ports

Here you can find the list with the default ports used by the Presto container.

<!--codeinclude-->
[Container Ports](../../modules/presto/presto.go) inside_block:containerPorts
<!--/codeinclude-->

### Container Options

When starting the Presto container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Presto container exposes the following methods:

#### ConnectionString

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the HTTP connection string for the Presto coordinator, e.g. `http://localhost:8080`.

```golang
connStr, err := prestoContainer.ConnectionString(ctx)
```
