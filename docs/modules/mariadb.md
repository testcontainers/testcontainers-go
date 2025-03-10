# MariaDB

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

## Introduction

The Testcontainers module for MariaDB.

## Adding this module to your project dependencies

Please run the following command to add the MariaDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mariadb
```

## Usage example

<!--codeinclude-->
[Creating a MariaDB container](../../modules/mariadb/examples_test.go) inside_block:runMariaDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The MariaDB module exposes one entrypoint function to create the MariaDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MariaDBContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the MariaDB container, you can pass options in a variadic way to configure it.

!!!tip

    You can find all the available configuration and environment variables for the MariaDB Docker image on [Docker Hub](https://hub.docker.com/_/mariadb).

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "mariadb:11.0.3")`.

!!!info
    From MariaDB [docs](https://github.com/docker-library/docs/tree/master/mariadb#environment-variables):

    From tag 10.2.38, 10.3.29, 10.4.19, 10.5.10 onwards, and all 10.6 and later tags,
    the `MARIADB_*` equivalent variables are provided. `MARIADB_*` variants will always be
    used in preference to `MYSQL_*` variants.

The MariaDB module will take all the environment variables that start with `MARIADB_` and duplicate them with the `MYSQL_` prefix.

{% include "../features/common_functional_options.md" %}

#### Set username, password and database name

If you need to set a different database, and its credentials, you can use `WithUsername`, `WithPassword`, `WithDatabase`
options.

!!!info
    The default values for the username is `root`, for password is `test` and for the default database name is `test`.

#### Init Scripts

If you would like to perform DDL or DML operations in the MariaDB container, add one or more `*.sql`, `*.sql.gz`, or `*.sh`
scripts to the container request, using the `WithScripts(scriptPaths ...string)`. Those files will be copied under `/docker-entrypoint-initdb.d`.

<!--codeinclude-->
[Example of Init script](../../modules/mariadb/testdata/schema.sql)
<!--/codeinclude-->

#### Custom configuration

If you need to set a custom configuration, you can use `WithConfigFile` option to pass the path to a custom configuration file.

### Container Methods

The MariaDB container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the MariaDB container, using the default `3306` port.
It's possible to pass extra parameters to the connection string, e.g. `tls=false`, in a variadic way.

!!!info
    By default, MariaDB transmits data between the server and clients without encrypting it.

<!--codeinclude-->
[Get connection string](../../modules/mariadb/mariadb_test.go) inside_block:connectionString
<!--/codeinclude-->
