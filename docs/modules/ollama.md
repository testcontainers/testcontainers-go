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

#### With Models

It's possible to initialise the Ollama container with a specific model passed as parameter. The supported models are described in the Ollama project: [https://github.com/ollama/ollama?tab=readme-ov-file](https://github.com/ollama/ollama?tab=readme-ov-file) and [https://ollama.com/library](https://ollama.com/library).

!!!warning
    At the moment you use one of those models, the Ollama image will load the model and could take longer to start because of that.

The following examples use the `llama2` model to connect to the Ollama container using HTTP and Langchain.

<!--codeinclude-->
[Using HTTP](../../modules/ollama/examples_test.go) inside_block:withHTTPModelLlama2
[Using Langchaingo](../../modules/ollama/examples_test.go) inside_block:withLangchainModelLlama2
<!--/codeinclude-->

### Container Methods

The Ollama container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the Ollama container, using the default `11434` port.

<!--codeinclude-->
[Get connection string](../../modules/ollama/ollama_test.go) inside_block:connectionString
<!--/codeinclude-->