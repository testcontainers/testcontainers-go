# Valkey

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Valkey.

## Adding this module to your project dependencies

Please run the following command to add the Valkey module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/valkey
```

## Usage example

<!--codeinclude-->
[Creating a Valkey container](../../modules/valkey/examples_test.go) inside_block:runValkeyContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Valkey module exposes one entrypoint function to create the Valkey container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ValkeyContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Valkey container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Valkey Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "valkey/valkey:7.2.5")`.

{% include "../features/common_functional_options.md" %}

#### Snapshotting

By default Valkey saves snapshots of the dataset on disk, in a binary file called dump.rdb. You can configure Valkey to have it save the dataset every `N` seconds if there are at least `M` changes in the dataset. E.g. `WithSnapshotting(10, 1)`.

#### Log Level

You can easily set the valkey logging level. E.g. `WithLogLevel(LogLevelDebug)`.

#### Valkey configuration

In the case you have a custom config file for Valkey, it's possible to copy that file into the container before it's started. E.g. `WithConfigFile(filepath.Join("testdata", "valkey.conf"))`.

### Container Methods

The Valkey container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the Valkey container, using the default `6379` port.

<!--codeinclude-->
[Get connection string](../../modules/valkey/valkey_test.go) inside_block:connectionString
<!--/codeinclude-->