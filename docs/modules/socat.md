# Socat

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Socat.

## Adding this module to your project dependencies

Please run the following command to add the Socat module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/socat
```

## Usage example

<!--codeinclude-->
[Creating a Socat container](../../modules/socat/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Socat module exposes one entrypoint function to create the Socat container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*SocatContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Socat container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "alpine/socat:1.8.0.1")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Socat container exposes the following methods:
