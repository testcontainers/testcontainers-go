# Minio

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

## Introduction

The Testcontainers module for Minio.

## Adding this module to your project dependencies

Please run the following command to add the Minio module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/minio
```

## Usage example

<!--codeinclude-->
[Creating a Minio container](../../modules/minio/examples_test.go) inside_block:runMinioContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Minio module exposes one entrypoint function to create the Minio container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MinioContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Minio container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "minio/minio:RELEASE.2024-01-16T16-07-38Z")`.

{% include "../features/common_functional_options.md" %}

#### Username and Password

If you need to set different credentials, you can use the `WithUsername(user string)` and `WithPassword(pwd string)` options.

### Container Methods

#### ConnectionString

This method returns the connection string to connect to the Minio container, using the default `9000` port.

<!--codeinclude-->
[Get connection string](../../modules/minio/minio_test.go) inside_block:connectionString
<!--/codeinclude-->
