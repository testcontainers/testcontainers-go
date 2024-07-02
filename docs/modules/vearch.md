# Vearch

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

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

## Module reference

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

If you need to set a different Vearch Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "vearch/vearch:3.5.1")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Vearch container exposes the following methods:

#### REST Endpoint

This method returns the REST endpoint of the Vearch container, using the default `9001` port.

<!--codeinclude-->
[Get REST endpoint](../../modules/vearch/vearch_test.go) inside_block:restEndpoint
<!--/codeinclude-->


