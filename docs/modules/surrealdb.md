# SurrealDB

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

## Introduction

The Testcontainers module for SurrealDB.

## Adding this module to your project dependencies

Please run the following command to add the SurrealDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/surrealdb
```

## Usage example

<!--codeinclude-->
[Creating a SurrealDB container](../../modules/surrealdb/examples_test.go) inside_block:runSurrealDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The SurrealDB module exposes one entrypoint function to create the SurrealDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*SurrealDBContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the SurrealDB container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "surrealdb/surrealdb:v1.1.1")`.

{% include "../features/common_functional_options.md" %}

#### Set username and password

If you need to set different credentials, you can use `WithUsername` and `WithPassword` options.

!!!info
    The default values for the username and the password are `root`.

#### WithAuthentication

If you need to enable authentication, you can use `WithAuthentication` option. By default, it is disabled.

#### WithStrictMode

If you need to enable the strict mode for SurrealDB, you can use `WithStrictMode` option. By default, it is disabled.

### WithAllowAllCaps

If you need to enable the all caps mode for SurrealDB, you can use `WithAllowAllCaps` option. By default, it is disabled.

### Container Methods

The SurrealDB container exposes the following methods:

#### URL

This method returns the websocket URL string to connect to the SurrealDB API, using the `8000` port.

<!--codeinclude-->
[Get websocket URL string](../../modules/surrealdb/surrealdb_test.go) inside_block:websocketURL
<!--/codeinclude-->
