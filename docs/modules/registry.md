# Registry

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Registry.

## Adding this module to your project dependencies

Please run the following command to add the Registry module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/registry
```

## Usage example

<!--codeinclude-->
[Creating a Registry container](../../modules/registry/examples_test.go) inside_block:runRegistryContainer
<!--/codeinclude-->

## Module reference

The Registry module exposes one entrypoint function to create the Registry container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RegistryContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Registry container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Registry Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Registry. E.g. `testcontainers.WithImage("registry:2.8.3")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Registry container exposes the following methods:
