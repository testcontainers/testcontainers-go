# YugabyteDB

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

## Introduction

The Testcontainers module for yugabyteDB.

## Adding this module to your project dependencies

Please run the following command to add the yugabyteDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/yugabytedb
```

## Usage example

<!--codeinclude-->
[Creating a yugabyteDB container](../../modules/yugabytedb/examples_test.go) inside_block:runyugabyteDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

The yugabyteDB module exposes one entrypoint function to create the yugabyteDB container, and this function receives three parameters:

```golang
func Run(
    ctx context.Context, 
    img string, 
    opts ...testcontainers.ContainerCustomizer,
) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the yugabyteDB container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "yugabytedb/yugabyte")`.

{% include "../features/common_functional_options.md" %}

#### Initial Database

By default the yugabyteDB container will start with a database named `yugabyte` and the default credentials `yugabyte` and `yugabyte`.

If you need to set a different database, and its credentials, you can use the `WithDatabaseName(dbName string)`, `WithDatabaseUser(dbUser string)` and `WithDatabasePassword(dbPassword string)` options.

#### Initial Cluster Configuration

By default the yugabyteDB container will start with a cluster keyspace named `yugabyte` and the default credentials `yugabyte` and `yugabyte`.

If you need to set a different cluster keyspace, and its credentials, you can use the `WithKeyspace(keyspace string)`, `WithUser(user string)` and `WithPassword(password string)` options.

### Container Methods

The yugabyteDB container exposes the following methods:

#### YSQLConnectionString

This method returns the connection string for the yugabyteDB container when using
the YSQL query language.
The connection string can then be used to connect to the yugabyteDB container using 
a standard PostgreSQL client.

<!--codeinclude-->
[Create a postgres client using the connection string](../../modules/yugabytedb/examples_test.go) block:ExampleContainer_YSQLConnectionString
<!--/codeinclude-->

### Usage examples

#### Usage with YSQL and gocql

To use the YCQL query language, you need to configure the cluster 
with the keyspace, user, and password.

By default, the yugabyteDB container will start with a cluster keyspace named `yugabyte` and the default credentials `yugabyte` and `yugabyte` but you can change it using the `WithKeyspace`, `WithUser` and `WithPassword` options.

In order to get the appropriate host and port to connect to the yugabyteDB container, 
you can use the `GetHost` and `GetMappedPort` methods on the Container struct.
See the examples below:

<!--codeinclude-->
[Create a yugabyteDB client using the cluster configuration](../../modules/yugabytedb/yugabytedb_test.go) block:TestYugabyteDB_YCQL
<!--/codeinclude-->