# Ollama

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Ollama.

## Adding this module to your project dependencies

Please run the following command to add the Ollama module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/ollama
```

## Usage example

<!--codeinclude-->
[Creating a Ollama container](../../modules/ollama/examples_test.go) inside_block:runOllamaContainer
<!--/codeinclude-->

## Module reference

The Ollama module exposes one entrypoint function to create the Ollama container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OllamaContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Ollama container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Ollama Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Ollama. E.g. `testcontainers.WithImage("ollama/ollama:0.1.25")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Ollama container exposes the following methods:
