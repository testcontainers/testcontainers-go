# CockroachDB

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

## Introduction

The Testcontainers module for CockroachDB.

## Adding this module to your project dependencies

Please run the following command to add the CockroachDB module to your Go dependencies:

```shell
go get github.com/testcontainers/testcontainers-go/modules/cockroachdb
```

## Usage example

<!--codeinclude-->
[Creating a CockroachDB container](../../modules/cockroachdb/examples_test.go) inside_block:runCockroachDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The CockroachDB module exposes one entrypoint function to create the CockroachDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the CockroachDB container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "cockroachdb/cockroach:latest-v23.1")`.

{% include "../features/common_functional_options.md" %}

#### Database

Set the database that is created & dialled with `cockroachdb.WithDatabase`.

#### User and Password

You can configure the container to create a user with a password by setting `cockroachdb.WithUser` and `cockroachdb.WithPassword`.

`cockroachdb.WithPassword` is incompatible with `cockroachdb.WithInsecure`.

#### Store size

Control the maximum amount of memory used for storage, by default this is 100% but can be changed by provided a valid option to `WithStoreSize`. Checkout https://www.cockroachlabs.com/docs/stable/cockroach-start#store for the full range of options available.

#### TLS authentication

`cockroachdb.WithInsecure` lets you disable the use of TLS on connections.

`cockroachdb.WithInsecure` is incompatible with `cockroachdb.WithPassword`.

#### Initialization Scripts

`cockroachdb.WithInitScripts` adds the given scripts to those automatically run when the container starts.
These will be ignored if data exists in the `/cockroach/cockroach-data` directory within the container.

`cockroachdb.WithNoClusterDefaults` disables the default cluster settings script.

Without this option Cockroach containers run `data/cluster-defaults.sql` on startup
which configures the settings recommended by Cockroach Labs for
[local testing clusters](https://www.cockroachlabs.com/docs/stable/local-testing)
unless data exists in the `/cockroach/cockroach-data` directory within the container.

### Container Methods

The CockroachDB container exposes the following methods:

#### ConnectionString

Dial address to open a new connection.

#### MustConnectionString

Same as `ConnectionString` but any error to generate the address will raise a panic

#### TLSConfig

Returns `*tls.Config` setup to allow you to dial your client over TLS, if enabled, else this will error with `cockroachdb.ErrTLSNotEnabled`.

!!!info
    The `TLSConfig()` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

#### ConnectionConfig

Returns `*pgx.ConnConfig` which can be passed to `pgx.ConnectConfig` to open a new connection.
