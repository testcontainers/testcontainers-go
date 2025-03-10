# Databend

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

## Introduction

The Testcontainers module for Databend.

## Adding this module to your project dependencies

Please run the following command to add the Databend module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/databend
```

## Usage example

<!--codeinclude-->
[Creating a Databend container](../../modules/databend/examples_test.go) inside_block:runDatabendContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

The Databend module exposes one entrypoint function to create the Databend container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DatabendContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Databend container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "datafuselabs/databend:v1.2.615")`.

{% include "../features/common_functional_options.md" %}

#### Set username, password

If you need to set a different user/password/database, you can use `WithUsername`, `WithPassword` options.

!!!info
The default values for the username is `databend`, for password is `databend` and for the default database name is `default`.

### Container Methods

The Databend container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the Databend container, using the default `8000` port.
It's possible to pass extra parameters to the connection string, e.g. `sslmode=disable`.

<!--codeinclude-->
[Get connection string](../../modules/databend/databend_test.go) inside_block:connectionString
<!--/codeinclude-->

#### MustGetConnectionString

`MustConnectionString` panics if the address cannot be determined.
