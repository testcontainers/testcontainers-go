# Minio

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

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

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Minio module exposes one entrypoint function to create the Minio container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MinioContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "quay.io/minio/minio:RELEASE.2024-01-16T16-07-38Z")`.

This Quay image requires registry authentication. Log in to `quay.io` with an account that has access to `minio/minio` before running the example or the module tests, or pass an image you can pull to `Run`. The Minio integration tests are excluded from CI until an accessible image is available.

### Container Options

When starting the Minio container, you can pass options in a variadic way to configure it.

#### Username and Password

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

If you need to set different credentials, you can use the `WithUsername(user string)` and `WithPassword(pwd string)` options.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

#### ConnectionString

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

This method returns the connection string to connect to the Minio container, using the default `9000` port.

<!--codeinclude-->
[Get connection string](../../modules/minio/minio_test.go) inside_block:connectionString
<!--/codeinclude-->
