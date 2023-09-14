# MySQL

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

## Introduction

The Testcontainers module for MySQL.

## Adding this module to your project dependencies

Please run the following command to add the MySQL module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mysql
```

## Usage example

<!--codeinclude--> 
[Creating a MySQL container](../../modules/mysql/examples_test.go) inside_block:runMySQLContainer
<!--/codeinclude-->

## Module Reference

The MySQL module exposes one entrypoint function to create the container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MySQLContainer, error) {
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the MySQL container, you can pass options in a variadic way to configure it.

!!!tip

    You can find all the available configuration and environment variables for the MySQL Docker image on [Docker Hub](https://hub.docker.com/_/mysql).

#### Image

If you need to set a different MySQL Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for MySQL. E.g. `testcontainers.WithImage("mysql:5.6")`.

By default, the container will use the following Docker image:

<!--codeinclude-->
[Default Docker image](../../modules/mysql/mysql.go) inside_block:defaultImage
<!--/codeinclude-->

#### Wait Strategies

If you need to set a different wait strategy for MySQL, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for MySQL.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for MySQL, you can leverage the following Docker type modifiers:

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

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the MySQL container is started.

#### Set username, password and database name

If you need to set a different database, and its credentials, you can use `WithUsername`, `WithPassword`, `WithDatabase`
options.

!!!info
    The default values for the username is `root`, for password is `test` and for the default database name is `test`.

#### Init Scripts

If you would like to perform DDL or DML operations in the MySQL container, add one or more `*.sql`, `*.sql.gz`, or `*.sh`
scripts to the container request, using the `WithScripts(scriptPaths ...string)`. Those files will be copied under `/docker-entrypoint-initdb.d`.

<!--codeinclude-->
[Example of Init script](../../modules/mysql/testdata/schema.sql)
<!--/codeinclude-->

#### Custom configuration

If you need to set a custom configuration, you can use `WithConfigFile` option to pass the path to a custom configuration file.

### Container Methods

#### ConnectionString

This method returns the connection string to connect to the MySQL container, using the default `3306` port.
It's possible to pass extra parameters to the connection string, e.g. `tls=skip-verify` or `application_name=myapp`, in a variadic way.

<!--codeinclude-->
[Get connection string](../../modules/mysql/mysql_test.go) inside_block:connectionString
<!--/codeinclude-->
