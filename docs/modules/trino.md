# Trino

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Trino.

[Trino](https://trino.io/) is an open-source distributed SQL query engine designed to run interactive analytic queries against data sources of all sizes, from gigabytes to petabytes. It is the community continuation of PrestoSQL.

## Adding this module to your project dependencies

Please run the following command to add the Trino module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/trino
```

## Usage example

<!--codeinclude-->
[Creating a Trino container](../../modules/trino/examples_test.go) inside_block:runTrinoContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Trino module exposes one entrypoint function to create the Trino container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "trinodb/trino:435")`.

### Wait strategy

The module waits for the Trino coordinator to be ready by polling the `/v1/info` HTTP endpoint on port `8080/tcp`. It checks that the JSON field `starting` is `false`, which indicates the coordinator has finished its initialisation. The default startup timeout is **2 minutes**.

### Container Options

When starting the Trino container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Trino container exposes the following methods:

#### ConnectionString

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the HTTP connection string for the Trino coordinator, e.g. `http://localhost:8080`.

```golang
connStr, err := trinoContainer.ConnectionString(ctx)
```
