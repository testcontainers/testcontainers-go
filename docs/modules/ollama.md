# Ollama

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

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

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Ollama module exposes one entrypoint function to create the Ollama container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OllamaContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Ollama container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Ollama Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "ollama/ollama:0.1.25")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Ollama container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the Ollama container, using the default `11434` port.

<!--codeinclude-->
[Get connection string](../../modules/ollama/ollama_test.go) inside_block:connectionString
<!--/codeinclude-->

#### Commit

This method commits the container to a new image, returning the new image ID.
It should be used after a model has been pulled and loaded into the container in order to create a new image with the model,
and eventually use it as the base image for a new container. That will speed up the execution of the following containers.

<!--codeinclude-->
[Commit Ollama image](../../modules/ollama/ollama_test.go) inside_block:commitOllamaContainer
<!--/codeinclude-->

## Examples

### Loading Models

It's possible to initialise the Ollama container with a specific model passed as parameter. The supported models are described in the Ollama project: [https://github.com/ollama/ollama?tab=readme-ov-file](https://github.com/ollama/ollama?tab=readme-ov-file) and [https://ollama.com/library](https://ollama.com/library).

!!!warning
    At the moment you use one of those models, the Ollama image will load the model and could take longer to start because of that.

The following examples use the `llama2` model to connect to the Ollama container using HTTP and Langchain.

<!--codeinclude-->
[Using HTTP](../../modules/ollama/examples_test.go) inside_block:withHTTPModelLlama2
[Using Langchaingo](../../modules/ollama/examples_test.go) inside_block:withLangchainModelLlama2
<!--/codeinclude-->
