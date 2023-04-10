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

The Postgres module exposes one entrypoint function to create the Postgres container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.CustomizeRequestOption) (*PostgresContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.CustomizeRequestOption`, a variadic argument for passing options.

### Container Options

When starting the Postgres container, you can pass options in a variadic way to configure it.

!!!tip
    You can find all the available configuration and environment variables for the Postgres Docker image on [Docker Hub](https://hub.docker.com/_/postgres).

#### Image

If you need to set a different Postgres Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Postgres. E.g. `testcontainers.WithImage("docker.io/postgres:9.6")`.

#### Initial Database

If you need to set a different database, and its credentials, you can use the `WithInitialDatabase`.

<!--codeinclude-->
[Set Initial database](../../modules/postgres/postgres_test.go) inside_block:withInitialDatabase
<!--/codeinclude-->

#### Init Scripts

If you would like to do additional initialization in the Postgres container, add one or more `*.sql`, `*.sql.gz`, or `*.sh` scripts to the container request.
Those files will be copied after the container is created but before it's started under `/docker-entrypoint-initdb.d`. According to Postgres Docker image,
it will run any `*.sql` files, run any executable `*.sh` scripts, and source any non-executable `*.sh` scripts found in that directory to do further
initialization before starting the service.

<!--codeinclude-->
[Include init scripts](../../modules/postgres/postgres_test.go) inside_block:withInitScripts
<!--/codeinclude-->

<!--codeinclude-->
[Init script content](../../modules/postgres/testdata/init-user-db.sh)
<!--/codeinclude-->

#### Database configuration

In the case you have a custom config file for Postgres, it's possible to copy that file into the container before it's started.

!!!tip
    For information on what is available to configure, see the [PostgreSQL docs](https://www.postgresql.org/docs/14/runtime-config.html) for the specific version of PostgreSQL that you are running. 

<!--codeinclude-->
[Include custom configuration file](../../modules/postgres/postgres_test.go) inside_block:withConfigFile
<!--/codeinclude-->

#### Wait Strategies

Given you could need to wait for different conditions, in particular using a wait.ForSQL strategy,
the Postgres container exposes a `testcontainers.WithWaitStrategy` option to set a custom wait strategy.

<!--codeinclude-->
[Set Wait Strategy](../../modules/postgres/postgres_test.go) inside_block:withWaitStrategy
<!--/codeinclude-->

### Container Methods

#### ConnectionString

This method returns the connection string to connect to the Postgres container, using the default `5432` port.
It's possible to pass extra parameters to the connection string, e.g. `sslmode=disable` or `application_name=myapp`, in a variadic way.

<!--codeinclude-->
[Get connection string](../../modules/postgres/postgres_test.go) inside_block:connectionString
<!--/codeinclude-->

### Postgres variants

It's possible to use the Postgres container with Timescale or Postgis, to name a few. You simply need to update the image name and the wait strategy.

<!--codeinclude-->
[Image for Timescale](../../modules/postgres/postgres_test.go) inside_block:timescale
<!--/codeinclude-->

<!--codeinclude-->
[Image for Postgis](../../modules/postgres/postgres_test.go) inside_block:postgis
<!--/codeinclude-->