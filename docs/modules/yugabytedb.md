# YugabyteDB

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

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
) (*yugabyteDBContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the yugabyteDB container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different yugabyteDB Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "yugabytedb/yugabyte")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The yugabyteDB container exposes the following methods:

#### YSQLConnectionString

This method returns the connection string for the yugabyteDB container when using
the YSQL query language.
The connection string can then be used to connect to the yugabyteDB container using 
a standard PostgreSQL client.

<!--codeinclude-->
[Create a postgres client using the connection string](../../modules/yugabytedb/examples_test.go) block:ExampleYugabyteDBContainer_YSQLConnectionString
<!--/codeinclude-->

#### YCQLConfigureClusterConfig

This method updates the passed in cluster configuration with the yugabyteDB container's
host, port, username and password information.
The cluster configuration can then be used to connect to the yugabyteDB container using
the official yugabyteDB Go client.

<!--codeinclude-->
[Create a yugabyteDB client using the cluster configuration](../../modules/yugabytedb/examples_test.go) block:ExampleYugabyteDBContainer_YCQLConfigureClusterConfig
<!--/codeinclude-->