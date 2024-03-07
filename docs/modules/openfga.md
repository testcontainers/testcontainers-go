# OpenFGA

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for OpenFGA.

## Adding this module to your project dependencies

Please run the following command to add the OpenFGA module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/openfga
```

## Usage example

<!--codeinclude-->
[Creating a OpenFGA container](../../modules/openfga/examples_test.go) inside_block:runOpenFGAContainer
<!--/codeinclude-->

## Module reference

The OpenFGA module exposes one entrypoint function to create the OpenFGA container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenFGAContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the OpenFGA container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different OpenFGA Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for OpenFGA. E.g. `testcontainers.WithImage("openfga/openfga:v1.5.0")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The OpenFGA container exposes the following methods:
