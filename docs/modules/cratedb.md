# CrateDB

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for CrateDB.

CrateDB is a distributed SQL database that combines the power of Lucene-based full-text search with the convenience of standard SQL. It offers PostgreSQL wire-protocol compatibility, making it easy to connect with existing PostgreSQL drivers and tools, and is optimised for real-time analytics on large datasets.

## Adding this module to your project dependencies

Please run the following command to add the CrateDB module to your Go dependencies:

```shell
go get github.com/testcontainers/testcontainers-go/modules/cratedb
```

## Usage example

<!--codeinclude-->
[Creating a CrateDB container](../../modules/cratedb/examples_test.go) inside_block:runCrateDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The CrateDB module exposes one entrypoint function to create the CrateDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Ports

Here you can find the list with the default ports used by the CrateDB container.

<!--codeinclude-->
[Container Ports](../../modules/cratedb/cratedb.go) inside_block:containerPorts
<!--/codeinclude-->

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "crate:5.7")`.

### Container Options

When starting the CrateDB container, you can pass options in a variadic way to configure it.

#### WithHeapSize

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the JVM heap size for the CrateDB process via the `CRATE_HEAP_SIZE` environment variable.
The default value is `512m`.

```golang
cratedb.WithHeapSize("1g")
```

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The CrateDB container exposes the following methods:

#### HTTPEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the HTTP endpoint (Admin UI and REST API) of the CrateDB container on port `4200/tcp`.

```golang
endpoint, err := cratedbContainer.HTTPEndpoint(ctx)
// endpoint: "http://localhost:32768"
```

#### PGConnectionString

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the PostgreSQL wire-protocol connection string for the CrateDB container on port `5432/tcp`.
CrateDB's built-in superuser is `crate` with no password, and the default schema is `doc`.
Accepts an optional variadic list of extra query parameters, e.g. `"sslmode=disable"` or `"connect_timeout=10"`.

```golang
connStr, err := cratedbContainer.PGConnectionString(ctx, "sslmode=disable")
// connStr: "postgres://crate@localhost:32769/doc?sslmode=disable"
```
