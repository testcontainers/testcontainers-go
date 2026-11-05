# CockroachDB

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

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

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

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

#### Database

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

Set the database that is created & dialled with `cockroachdb.WithDatabase`.

#### User and Password

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

You can configure the container to create a user with a password by setting `cockroachdb.WithUser` and `cockroachdb.WithPassword`.

`cockroachdb.WithPassword` is incompatible with `cockroachdb.WithInsecure`.

#### Store size

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

Control the maximum amount of memory used for storage, by default this is 100% but can be changed by provided a valid option to `WithStoreSize`. Checkout https://www.cockroachlabs.com/docs/stable/cockroach-start#store for the full range of options available.

#### TLS authentication

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.35.0"><span class="tc-version">:material-tag: v0.35.0</span></a>

`cockroachdb.WithInsecure` lets you disable the use of TLS on connections.

`cockroachdb.WithInsecure` is incompatible with `cockroachdb.WithPassword`.

#### Initialization Scripts

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.35.0"><span class="tc-version">:material-tag: v0.35.0</span></a>

`cockroachdb.WithInitScripts` adds the given scripts to those automatically run when the container starts.
These will be ignored if data exists in the `/cockroach/cockroach-data` directory within the container.

#### No Cluster Defaults

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.35.0"><span class="tc-version">:material-tag: v0.35.0</span></a>

`cockroachdb.WithNoClusterDefaults` disables the default cluster settings script.

Without this option Cockroach containers run `data/cluster-defaults.sql` on startup
which configures the settings recommended by Cockroach Labs for
[local testing clusters](https://www.cockroachlabs.com/docs/stable/local-testing)
unless data exists in the `/cockroach/cockroach-data` directory within the container.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The CockroachDB container exposes the following methods:

#### ConnectionString

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

Returns a portable PostgreSQL URL. In secure mode the URL does not contain the
in-memory CA or client certificate; use `ConnectionInfo` or the optional pgx
adapter below to configure TLS.


#### MustConnectionString

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

Same URL semantics as `ConnectionString`; panics if address lookup fails.

#### TLSConfig

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

Returns `*tls.Config` setup to allow you to dial your client over TLS, if enabled, else this will error with `cockroachdb.ErrTLSNotEnabled`.

!!!info
    The `TLSConfig()` function is deprecated. Use the `TLSConfig` field returned by
    `ConnectionInfo(ctx)` for a cloned TLS configuration.

#### ConnectionInfo

Returns `cockroachdb.ConnectionInfo` with `URL string` and `TLSConfig *tls.Config`.
The TLS config is cloned and is nil in insecure mode. Apply it to your client
when connecting securely; preserve its CA trust, client certificate and
`ServerName` (`127.0.0.1` by default).

### Breaking change: pgx adapter moved

!!! warning "Combined CockroachDB and Dex SDK extraction"
    `(*cockroachdb.CockroachDBContainer).ConnectionConfig(ctx)` has been removed
    from `github.com/testcontainers/testcontainers-go/modules/cockroachdb`.
    The pgx API now belongs to the importable child package
    `github.com/testcontainers/testcontainers-go/modules/cockroachdb/examples`.
    This change ships together with the [Dex admin adapter move](dex.md#breaking-change-admin-adapter-moved).

Import the adapter as `cockroachdbexamples` and wrap the parent container:

```go
import cockroachdbexamples "github.com/testcontainers/testcontainers-go/modules/cockroachdb/examples"

// Before: cfg, err := c.ConnectionConfig(ctx)
clientContainer := cockroachdbexamples.NewContainer(c)
cfg, err := clientContainer.ConnectionConfig(ctx) // *pgx.ConnConfig
// Pass cfg to pgx.ConnectConfig after checking err.

// Before: addr, err := c.ConnectionString(ctx)
addr, err := clientContainer.ConnectionString(ctx)
// Pass addr to sql.Open("pgx/v5", addr) after checking err.
```

Move explicit method expressions, interfaces and function references involving
`ConnectionConfig(context.Context) (*pgx.ConnConfig, error)` to the child
`*cockroachdbexamples.Container`. Its `ConnectionString` and
`MustConnectionString` preserve the old opaque `stdlib.RegisterConnConfig`
handle semantics. The adapter imports pgx's stdlib driver; callers own closing
pgx connections and `*sql.DB`, and can release unused registered handles with
`stdlib.UnregisterConnConfig`.

**The parent's `ConnectionString` and `MustConnectionString` now return URLs,
not registered handles**, despite their unchanged Go signatures. Existing
`sql.Open` callers must use the adapter to retain the in-memory certificate
configuration. A URL alone does not serialize the CA or client certificate.
Consumers implementing their own client must use `ConnectionInfo` and apply
its TLS settings, and register their own SQL driver when using `database/sql`.
The parent no longer registers `pgx/v5` as an import side effect.

Provisioning and options remain in the parent. Readiness still waits, in order,
for `init_success`, HTTP health, and (in secure mode) CA/client certificates,
then executes `cockroach sql --url <connection URL> --execute 'SELECT 1'`
inside the container using the configured database, user and password. Secure
mode uses the container's CA/client certificate/key and `127.0.0.1`; insecure
mode uses `sslmode=disable`. A nonzero SQL exit retries until the existing
startup timeout or caller context expires. This replaces the external pgx SQL
probe; applications still own connectivity from their process to the mapped
endpoint.

The adapter is optional example support in a separate Go module. Starting with
this breaking release, the release script tags and publishes the parent and
adapter modules at the same version. Set `VERSION` to the released version you
are upgrading to and install both:

```sh
go get github.com/testcontainers/testcontainers-go/modules/cockroachdb@"${VERSION}" github.com/testcontainers/testcontainers-go/modules/cockroachdb/examples@"${VERSION}"
```

To consume an unreleased local checkout, add replacements for both modules (substitute
your checkout path), then add the adapter:

```sh
go mod edit -replace=github.com/testcontainers/testcontainers-go/modules/cockroachdb=/path/to/testcontainers-go/modules/cockroachdb
go mod edit -replace=github.com/testcontainers/testcontainers-go/modules/cockroachdb/examples=/path/to/testcontainers-go/modules/cockroachdb/examples
go get github.com/testcontainers/testcontainers-go/modules/cockroachdb/examples
```

Run the SDK tests and both runnable client examples from
`modules/cockroachdb/examples`; run the independent parent suite from
`modules/cockroachdb`. Parent `go test ./...` does not traverse the nested
module. The child pins pgx/v5 at `v5.9.2`.
