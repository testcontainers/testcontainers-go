# Postgres

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

## Introduction

The Testcontainers module for Postgres.

## Adding this module to your project dependencies

Please run the following command to add the Postgres module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

## Usage example

<!--codeinclude-->
[Creating a Postgres container](../../modules/postgres/examples_test.go) inside_block:runPostgresContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Postgres module exposes one entrypoint function to create the Postgres container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*PostgresContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Postgres container, you can pass options in a variadic way to configure it.

!!!tip
    You can find all the available configuration and environment variables for the Postgres Docker image on [Docker Hub](https://hub.docker.com/_/postgres).

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "postgres:16-alpine")`.

{% include "../features/common_functional_options.md" %}

#### Initial Database

If you need to set a different database, and its credentials, you can use the `WithDatabase(db string)`, `WithUsername(user string)` and `WithPassword(pwd string)` options.

#### Init Scripts

If you would like to do additional initialization in the Postgres container, add one or more `*.sql`, `*.sql.gz`, or `*.sh` scripts to the container request with the `WithInitScripts` function.
Those files will be copied after the container is created but before it's started under `/docker-entrypoint-initdb.d`. According to Postgres Docker image,
it will run any `*.sql` files, run any executable `*.sh` scripts, and source any non-executable `*.sh` scripts found in that directory to do further
initialization before starting the service.

An example of a `*.sh` script that creates a user and database is shown below:

<!--codeinclude-->
[Init script content](../../modules/postgres/testdata/init-user-db.sh)
<!--/codeinclude-->

#### Database configuration

In the case you have a custom config file for Postgres, it's possible to copy that file into the container before it's started, using the `WithConfigFile(cfgPath string)` function.

This function can be used `WithSSLSettings` but requires your configuration correctly sets the SSL properties. See the below section for more information.

!!!tip
    For information on what is available to configure, see the [PostgreSQL docs](https://www.postgresql.org/docs/14/runtime-config.html) for the specific version of PostgreSQL that you are running.

#### SSL Configuration

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.35.0"><span class="tc-version">:material-tag: v0.35.0</span></a>

If you would like to use SSL with the container you can use the `WithSSLSettings`. This function accepts a `SSLSettings` which has the required secret material, namely the ca-certificate, server certificate and key. The container will copy this material to `/tmp/testcontainers-go/postgres/ca_cert.pem`, `/tmp/testcontainers-go/postgres/server.cert` and `/tmp/testcontainers-go/postgres/server.key`

This function requires a custom postgres configuration file that enables SSL and correctly sets the paths on the key material.

If you use this function by itself or in conjunction with `WithConfigFile` your custom conf must set the required ssl fields. The configuration must correctly align the key material provided via `SSLSettings` with the server configuration, namely the paths. Your configuration will need to contain the following:

```
ssl = on
ssl_ca_file = '/tmp/testcontainers-go/postgres/ca_cert.pem'
ssl_cert_file = '/tmp/testcontainers-go/postgres/server.cert'
ssl_key_file = '/tmp/testcontainers-go/postgres/server.key'
```

!!!warning
    This function assumes the postgres user in the container is `postgres`

    There is no current support for mutual authentication.

    The `SSLSettings` function will modify the container `entrypoint`. This is done so that key material copied over to the container is chowned by `postgres`. All other container arguments will be passed through to the original container entrypoint.

### Container Methods

#### ConnectionString

This method returns the connection string to connect to the Postgres container, using the default `5432` port.
It's possible to pass extra parameters to the connection string, e.g. `sslmode=disable` or `application_name=myapp`, in a variadic way.

<!--codeinclude-->
[Get connection string](../../modules/postgres/postgres_test.go) inside_block:connectionString
<!--/codeinclude-->

### Postgres variants

It's possible to use the Postgres container with PGVector, Timescale or Postgis, to name a few. You simply need to update the image name and the wait strategy.

<!--codeinclude-->
[Image for PGVector](../../modules/postgres/postgres_test.go) inside_block:pgvector
[Image for Timescale](../../modules/postgres/postgres_test.go) inside_block:timescale
[Image for Postgis](../../modules/postgres/postgres_test.go) inside_block:postgis
<!--/codeinclude-->

## Examples

### Wait Strategies

The postgres module works best with these wait strategies.
No default is supplied, so you need to set it explicitly.

<!--codeinclude-->
[Example Wait Strategies](../../modules/postgres/wait_strategies.go) inside_block:waitStrategy
<!--/codeinclude-->

### Using Snapshots
This example shows the usage of the postgres module's Snapshot feature to give each test a clean database without having
to recreate the database container on every test or run heavy scripts to clean your database. This makes the individual
tests very modular, since they always run on a brand-new database.

!!!tip
    You should never pass the `"postgres"` system database as the container database name if you want to use snapshots. 
    The Snapshot logic requires dropping the connected database and using the system database to run commands, which will
    not work if the database for the container is set to `"postgres"`.

<!--codeinclude-->
[Test with a reusable Postgres container](../../modules/postgres/postgres_test.go) inside_block:snapshotAndReset
<!--/codeinclude-->

### Snapshot/Restore with custom driver

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

The snapshot/restore feature tries to use the `postgres` driver with go's included `sql.DB` package to perform database operations.
If the `postgres` driver is not installed, it will fall back to using `docker exec`, which works, but is slower.

You can tell the module to use the database driver you have imported in your test package by setting `postgres.WithSQLDriver("name")` to your driver name.

For example, if you use pgx, see the example below.

```go
package my_test

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)
```

The above code snippet is importing the `pgx` driver and the _Testcontainers for Go_ Postgres module.

<!--codeinclude-->
[Snapshot/Restore with custom driver](../../modules/postgres/postgres_test.go) inside_block:snapshotAndReset
<!--/codeinclude-->