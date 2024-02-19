# Weaviate

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Weaviate.

## Adding this module to your project dependencies

Please run the following command to add the Weaviate module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/weaviate
```

## Usage example

<!--codeinclude-->
[Creating a Weaviate container](../../modules/weaviate/examples_test.go) inside_block:runWeaviateContainer
<!--/codeinclude-->

## Module reference

The Weaviate module exposes one entrypoint function to create the Weaviate container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*WeaviateContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Weaviate container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Weaviate Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Weaviate. E.g. `testcontainers.WithImage("semitechnologies/weaviate:1.23.9")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Weaviate container exposes the following methods:
