# Minio

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

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

## Module reference

The Minio module exposes one entrypoint function to create the Minio container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MinioContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Minio container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Minio Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Minio. E.g. `testcontainers.WithImage("minio/minio:RELEASE.2024-01-16T16-07-38Z")`.

{% include "../features/common_functional_options.md" %}

#### Username and Password

If you need to set different credentials, you can use the `WithUsername(user string)` and `WithPassword(pwd string)` options.

### Container Methods

#### ConnectionString

This method returns the connection string to connect to the Minio container, using the default `9000` port.

<!--codeinclude-->
[Get connection string](../../modules/minio/minio_test.go) inside_block:connectionString
<!--/codeinclude-->
