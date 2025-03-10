# Vearch

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

## Introduction

The Testcontainers module for Vearch.

## Adding this module to your project dependencies

Please run the following command to add the Vearch module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/vearch
```

## Usage example

<!--codeinclude-->
[Creating a Vearch container](../../modules/vearch/examples_test.go) inside_block:runVearchContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Vearch module exposes one entrypoint function to create the Vearch container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*VearchContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Vearch container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "vearch/vearch:3.5.1")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Vearch container exposes the following methods:

#### REST Endpoint

This method returns the REST endpoint of the Vearch container, using the default `9001` port.

<!--codeinclude-->
[Get REST endpoint](../../modules/vearch/vearch_test.go) inside_block:restEndpoint
<!--/codeinclude-->


