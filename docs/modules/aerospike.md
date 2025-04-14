# Aerospike

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Aerospike.

## Adding this module to your project dependencies

Please run the following command to add the Aerospike module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/aerospike
```

## Usage example

<!--codeinclude-->
[Creating a Aerospike container](../../modules/aerospike/examples_test.go) inside_block:runAerospikeContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Aerospike module exposes one entrypoint function to create the Aerospike container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*AerospikeContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Aerospike container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "aerospike/aerospike-server:latest")`.

{% include "../features/common_functional_options.md" %}

## Examples

### Aerospike Client

<!--codeinclude-->
[Aerospike Client](../../modules/aerospike/examples_test.go) inside_block:usingClient
<!--/codeinclude-->
