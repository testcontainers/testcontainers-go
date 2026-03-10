# TiDB

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for TiDB. TiDB is a MySQL-compatible distributed SQL database.

## Adding this module to your project dependencies

Please run the following command to add the TiDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/tidb
```

## Usage example

<!--codeinclude--> 
[Creating a TiDB container](../../modules/tidb/examples_test.go) inside_block:runTiDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

The TiDB module exposes one entrypoint function to create the container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "pingcap/tidb:v8.4.0")`.

### Container Options

When starting the TiDB container, you can pass options in a variadic way to configure it.

!!!tip
    You can find all the available configuration and environment variables for the TiDB Docker image on [Docker Hub](https://hub.docker.com/r/pingcap/tidb).

!!!info
    TiDB is MySQL-compatible, so you can use any MySQL driver (e.g. `github.com/go-sql-driver/mysql`) to connect and execute SQL statements after the container is ready.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

#### ConnectionString

This method returns the connection string to connect to the TiDB container, using the default `4000` port.
It's possible to pass extra parameters to the connection string, e.g. `charset=utf8mb4`, in a variadic way.

<!--codeinclude-->
[Get connection string](../../modules/tidb/tidb_test.go) inside_block:connectionString
<!--/codeinclude-->

#### MustConnectionString

Same as `ConnectionString`, but panics if an error occurs while getting the connection string.
