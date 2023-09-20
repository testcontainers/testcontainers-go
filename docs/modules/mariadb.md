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

## Module reference

The MariaDB module exposes one entrypoint function to create the MariaDB container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MariaDBContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the MariaDB container, you can pass options in a variadic way to configure it.

!!!tip

    You can find all the available configuration and environment variables for the MariaDB Docker image on [Docker Hub](https://hub.docker.com/_/mariadb).

#### Image

If you need to set a different MariaDB Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for MariaDB. E.g. `testcontainers.WithImage("mariadb:11.0.3")`.

!!!info
    From MariaDB [docs](https://github.com/docker-library/docs/tree/master/mariadb#environment-variables):

    From tag 10.2.38, 10.3.29, 10.4.19, 10.5.10 onwards, and all 10.6 and later tags,
    the `MARIADB_*` equivalent variables are provided. `MARIADB_*` variants will always be
    used in preference to `MYSQL_*` variants.

The MariaDB module will take all the environment variables that start with `MARIADB_` and duplicate them with the `MYSQL_` prefix.

#### Wait Strategies

If you need to set a different wait strategy for MariaDB, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for MariaDB.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for MariaDB, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

#### Startup Commands

!!!info
    Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Testcontainers exposes the `WithStartupCommand(e ...Executable)` option to run arbitrary commands in the container right after it's started.

!!!info
    To better understand how this feature works, please read the [Create containers: Lifecycle Hooks](../../features/creating_container/#lifecycle-hooks) documentation.

It also exports an `Executable` interface, defining one single method: `AsCommand()`, which returns a slice of strings to represent the command and positional arguments to be executed in the container.

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the MariaDB container is started.

#### Set username, password and database name

If you need to set a different database, and its credentials, you can use `WithUsername`, `WithPassword`, `WithDatabase`
options.

!!!info
    The default values for the username is `root`, for password is `test` and for the default database name is `test`.

#### Init Scripts

If you would like to perform DDL or DML operations in the MariaDB container, add one or more `*.sql`, `*.sql.gz`, or `*.sh`
scripts to the container request. Those files will be copied under `/docker-entrypoint-initdb.d`.

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
