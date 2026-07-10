# KurrentDB

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [KurrentDB](https://www.kurrent.io/kurrentdb), an event-native database for storing application state changes as immutable events. KurrentDB is the rebrand of EventStoreDB and uses the `kurrent://` connection string scheme.

## Adding this module to your project dependencies

Please run the following command to add the KurrentDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/kurrentdb
```

## Usage example

<!--codeinclude-->
[Creating a KurrentDB container](../../modules/kurrentdb/examples_test.go) inside_block:runKurrentDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The KurrentDB module exposes one entrypoint function to create the KurrentDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "kurrentplatform/kurrentdb:latest")`.

### Container Options

When starting the KurrentDB container, you can pass options in a variadic way to configure it.

#### WithInsecure

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

By default, KurrentDB runs in insecure mode (no TLS), which is suitable for local development and testing. The `WithInsecure` option makes this explicit:

<!--codeinclude-->
[Run with insecure mode](../../modules/kurrentdb/examples_test.go) inside_block:withInsecure
<!--/codeinclude-->

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The KurrentDB container exposes the following methods:

#### ConnectionString

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `ConnectionString(ctx)` method returns the connection string to connect to the KurrentDB container using the `kurrent://` scheme.
When the container runs in insecure mode, `?tls=false` is appended.

<!--codeinclude-->
[Get connection string](../../modules/kurrentdb/examples_test.go) inside_block:connectionString
<!--/codeinclude-->
