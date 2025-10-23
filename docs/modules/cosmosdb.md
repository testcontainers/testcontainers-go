# CosmosDB

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for CosmosDB.

## Adding this module to your project dependencies

Please run the following command to add the CosmosDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/cosmosdb
```

## Usage example

<!--codeinclude-->
[Creating a CosmosDB container](../../modules/cosmosdb/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The CosmosDB module exposes one entrypoint function to create the CosmosDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:vnext-preview")`.

### Container Options

When starting the CosmosDB container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The CosmosDB container exposes the following methods:
