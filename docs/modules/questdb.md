# QuestDB

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

QuestDB is an open-source time-series database designed for high-throughput ingestion and fast SQL queries. It supports the InfluxDB line protocol for fast writes, a PostgreSQL wire protocol for SQL queries, and a REST API with a built-in web console.

## Adding this module to your project dependencies

Please run the following command to add the QuestDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/questdb
```

## Usage example

<!--codeinclude-->
[Creating a QuestDB container](../../modules/questdb/examples_test.go) inside_block:runQuestDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The QuestDB module exposes one entrypoint function to create the QuestDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "questdb/questdb:7.4")`.

### Container Options

When starting the QuestDB container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The QuestDB container exposes the following methods:

#### HTTPEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the URL of the QuestDB HTTP web console and REST API, using the `9000/tcp` port.

<!--codeinclude-->
[HTTPEndpoint](../../modules/questdb/examples_test.go) inside_block:httpEndpoint
<!--/codeinclude-->

The returned URL has the format `http://host:port`.

#### PGEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the PostgreSQL wire protocol connection string for the QuestDB container, using the `8812/tcp` port and the built-in default credentials (`admin`/`quest`).

<!--codeinclude-->
[PGEndpoint](../../modules/questdb/examples_test.go) inside_block:pgEndpoint
<!--/codeinclude-->

The returned URL has the format `postgres://admin:[REDACTED]@host:port/qdb`.

#### InfluxDBEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the InfluxDB line protocol endpoint address of the QuestDB container, using the `9009/tcp` port. This endpoint is used for high-throughput time-series data ingestion.

<!--codeinclude-->
[InfluxDBEndpoint](../../modules/questdb/examples_test.go) inside_block:influxDBEndpoint
<!--/codeinclude-->

The returned address has the format `host:port`.
