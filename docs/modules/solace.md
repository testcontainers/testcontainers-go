# Solace Pubsub+

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Solace Pubsub+.

## Adding this module to your project dependencies

Please run the following command to add the Solace Pubsub+ module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/solace
```

## Usage example

<!--codeinclude-->
[Creating a Solace Pubsub+ container](../../modules/solace/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Solace Pubsub+ module exposes one entrypoint function to create the Solace Pubsub+ container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "solace-pubsub-standard:latest")`.

### Container Options

When starting the Solace Pubsub+ container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Solace Pubsub+ container exposes the following methods:
