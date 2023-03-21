# Postgres

## Adding this module to your project dependencies

Please run the following command to add the Postgres module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

## Usage example

<!--codeinclude-->
[Creating a Postgres container](../../modules/postgres/postgres_test.go) inside_block:postgresCreateContainer
<!--/codeinclude-->

## Module reference

The Postgres module exposes one single function to create the Postgres container, and this function receives two parameters:

```golang
func StartContainer(ctx context.Context, opts ...PostgresContainerOption) (*PostgresContainer, error)
```

- `context.Context`, the Go context.
- `PostgresContainerOption`, a variad argument for passing options.

## Container Options

When starting the Postgres container, you can pass options in a variadic way to configure it.

### Image

If you need to set a different Postgres Docker image, you can use `WithImage` with a valid Docker image
for Postgres. E.g. `WithImage("docker.io/postgres:9.6")`.

### Initial Database

If you need to set a different database, and its credentials, you can use the `WithInitialDatabase`.

<!--codeinclude-->
[Set Initial database](../../modules/postgres/postgres_test.go) inside_block:withInitialDatabase
<!--/codeinclude-->

### Wait Strategies

Given you could need to wait for different conditions, in particular using a wait.ForSQL strategy,
the Postgres container exposes a `WithWaitStrategy` option to set a custom wait strategy.

<!--codeinclude-->
[Set Wait Strategy](../../modules/postgres/postgres_test.go) inside_block:withWaitStrategy
<!--/codeinclude-->

### InitDB arguments

This optional function can be used to send arguments to postgres initdb. The value is a space separated string of arguments as postgres initdb would expect them.
This is useful for adding functionality like data page checksums: `WithInitDBArgs("--data-checksums")`.
