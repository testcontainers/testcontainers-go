# FakeSnow

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

[fakesnow](https://github.com/tekumara/fakesnow) is a Snowflake fake backed by DuckDB. It runs as an
HTTP server speaking the Snowflake REST API, so the official
[gosnowflake](https://github.com/snowflakedb/gosnowflake) driver connects to it unmodified and
Snowflake-flavoured SQL such as `VARIANT`, `LATERAL FLATTEN` and `QUALIFY` works without a Snowflake account.

## Adding this module to your project dependencies

Please run the following command to add the FakeSnow module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/fakesnow
```

## Usage example

<!--codeinclude-->
[Creating a FakeSnow container](../../modules/fakesnow/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

Connecting to it with `database/sql`:

<!--codeinclude-->
[Querying a FakeSnow container](../../modules/fakesnow/examples_test.go) inside_block:ExampleContainer_ConnectionString
<!--/codeinclude-->

fakesnow starts with no databases, so create the one the DSN points at before using it.

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The FakeSnow module exposes one entrypoint function to create the FakeSnow container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "ghcr.io/tekumara/fakesnow:0.11.21")`.

### Container Options

When starting the FakeSnow container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

#### Connection details

fakesnow accepts any credentials, so these options only change what the driver sends.

- `WithAccount(account string)`, the account, `fakesnow` by default. It cannot be empty: gosnowflake
  then derives the account from the host, which has no account part when the host is a container address.
- `WithUsername(user string)`, the user, `fakesnow` by default.
- `WithPassword(password string)`, the password, `fakesnow` by default.
- `WithDatabase(database string)`, the database in the DSN, `db1` by default. It is not created for you.
- `WithSchema(schema string)`, the schema in the DSN, `schema1` by default. It is not created for you.

### Container Methods

The FakeSnow container exposes the following methods:

#### ConnectionString

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns a gosnowflake DSN for the container, for use with `sql.Open("snowflake", dsn)`.

<!--codeinclude-->
[Get connection string](../../modules/fakesnow/fakesnow.go) inside_block:ConnectionString
<!--/codeinclude-->
