# Memcached

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.38.0"><span class="tc-version">:material-tag: v0.38.0</span></a>

## Introduction

The Testcontainers module for Memcached.

## Adding this module to your project dependencies

Please run the following command to add the Memcached module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/memcached
```

## Usage example

<!--codeinclude-->
[Creating a Memcached container](../../modules/memcached/examples_test.go) inside_block:runMemcachedContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.38.0"><span class="tc-version">:material-tag: v0.38.0</span></a>

The Memcached module exposes one entrypoint function to create the Memcached container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "memcached:1.6-alpine")`.

### Container Options

When starting the Memcached container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Memcached container exposes the following methods:

#### HostPort

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.38.0"><span class="tc-version">:material-tag: v0.38.0</span></a>

The `HostPort` method returns the host and port of the Memcached container, in the format `host:port`. Use this method to connect to the Memcached container from your application.

<!--codeinclude-->
[Get host:port](../../modules/memcached/examples_test.go) inside_block:hostPort
<!--/codeinclude-->
