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

{% include "../features/common_functional_options.md" %}

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
