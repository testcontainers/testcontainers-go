# MySQL

## Adding this module to your project dependencies

Please run the following command to add the MySQL module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mysql
```

## Usage example

<!--codeinclude--> 
[Creating a MySQL container](../../modules/mysql/mysql_test.go) inside_block:createMysqlContainer
<!--/codeinclude-->

## Module Reference

The MySQL module exposes one entrypoint function to create the container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.CustomizeRequestOption) (*MySQLContainer, error) {
```

- `context.Context`, the Go context.
- `testcontainers.CustomizeRequestOption`, a variad argument for passing options.

## Container Options

When starting the MySQL container, you can pass options in a variadic way to configure it.

!!!tip

    You can find all the available configuration and environment variables for the MySQL Docker image on [Docker Hub](https://hub.docker.com/_/mysql).

### Set Image

By default, the image used is `mysql:8`.  If you need to use a different image, you can use `testcontainers.WithImage` option.

<!--codeinclude-->
[Custom Image](../../modules/mysql/mysql_test.go) inside_block:withConfigFile
<!--/codeinclude-->

### Set username, password and database name

If you need to set a different database, and its credentials, you can use `WithUsername`, `WithPassword`, `WithDatabase`
options.  By default, the username, the password and the database name is `test`.

<!--codeinclude-->
[Custom Database initialization](../../modules/mysql/mysql_test.go) inside_block:customInitialization
<!--/codeinclude-->

### Init Scripts

If you would like to perform DDL or DML operations in the MySQL container, add one or more `*.sql`, `*.sql.gz`, or `*.sh`
scripts to the container request. Those files will be copied under `/docker-entrypoint-initdb.d`.

<!--codeinclude-->
[Include init scripts](../../modules/mysql/mysql_test.go) inside_block:withScripts
<!--/codeinclude-->

### Custom configuration

If you need to set a custom configuration, you can use `WithConfigFile` option.

<!--codeinclude-->
[Custom MySQL config file](../../modules/mysql/mysql_test.go) inside_block:withConfigFile
<!--/codeinclude-->
