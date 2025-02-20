# Azurite

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

## Introduction

The Testcontainers module for Azurite.

## Adding this module to your project dependencies

Please run the following command to add the Azurite module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/azurite
```

## Usage example

<!--codeinclude-->
[Creating an Azurite container](../../modules/azurite/examples_test.go) inside_block:runAzuriteContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Azurite module exposes one entrypoint function to create the Azurite container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*AzuriteContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Default Credentials

The Azurite container uses the following default credentials:

<!--codeinclude-->
[Default Credentials](../../modules/azurite/azurite.go) inside_block:defaultCredentials
<!--/codeinclude-->

### Container Options

When starting the Azurite container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Azurite Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "mcr.microsoft.com/azure-storage/azurite:3.28.0")`.

{% include "../features/common_functional_options.md" %}

#### WithInMemoryPersistence

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

If you want to use in-memory persistence, you can use `WithInMemoryPersistence(megabytes float64)`. E.g. `azurite.WithInMemoryPersistence(64.0)`.

Please read the [Azurite documentation](https://github.com/Azure/Azurite?tab=readme-ov-file#use-in-memory-storage) for more information.

!!! warning
    This option is only available in Azurite versions 3.28.0 and later.

### Container Methods

The Azurite container exposes the following methods:

#### ServiceURL

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

Returns the service URL to connect to the Azurite container and an error, passing the Go context and the service name as parameters.

#### MustServiceURL

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

Returns the service URL to connect to the Azurite container, passing the Go context and the service name as parameters. If an error occurs, it will panic.

### Examples

#### Blob Operations

In the following example, we will create a container with Azurite and perform some blob operations. For that, using the default
credentials, we will create an Azurite container, upload a blob to it, list the blobs, and download the blob. Finally, we will remove the created blob and container.

<!--codeinclude-->
[Performing blob operations](../../modules/azurite/examples_test.go) inside_block:blobOperations
<!--/codeinclude-->

#### Queue Operations

In the following example, we will create an Azurite container and perform some queue operations. For that, using the default
credentials, we will create a queue, list the queues, and finally we will remove the created queue.

<!--codeinclude-->
[Performing queue operations](../../modules/azurite/examples_test.go) inside_block:queueOperations
<!--/codeinclude-->

#### Table Operations

In the following example, we will create an Azurite container and perform some table operations. For that, using the default
credentials, we will create a table, list the tables, and finally we will remove the created table.

<!--codeinclude-->
[Performing table operations](../../modules/azurite/examples_test.go) inside_block:tableOperations
<!--/codeinclude-->