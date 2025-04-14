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

The module allows you to run the Ollama container or the local Ollama binary.

<!--codeinclude-->
[Creating an Ollama container](../../modules/ollama/examples_test.go) inside_block:runOllamaContainer
[Running the local Ollama binary](../../modules/ollama/examples_test.go) inside_block:localOllama
<!--/codeinclude-->

If the local Ollama binary fails to execute, the module will fall back to the container version of Ollama.

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

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "ollama/ollama:0.5.7")`.

#### Use Local

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.35.0"><span class="tc-version">:material-tag: v0.35.0</span></a>

!!!warning
    Please make sure the local Ollama binary is not running when using the local version of the module:
    Ollama can be started as a system service, or as part of the Ollama application,
    and interacting with the logs of a running Ollama process not managed by the module is not supported.

If you need to run the local Ollama binary, you can set the `UseLocal` option in the `Run` function.
This option accepts a list of environment variables as a string, that will be applied to the Ollama binary when executing commands.

E.g. `Run(context.Background(), "ollama/ollama:0.5.7", WithUseLocal("OLLAMA_DEBUG=true"))`.

All the container methods are available when using the local Ollama binary, but will be executed locally instead of inside the container.
Please consider the following differences when using the local Ollama binary:

- The local Ollama binary will create a log file in the current working directory, identified by the session ID. E.g. `local-ollama-<session-id>.log`. It's possible to set the log file name using the `OLLAMA_LOGFILE` environment variable. So if you're running Ollama yourself, from the Ollama app, or the standalone binary, you could use this environment variable to set the same log file name.
  - For the Ollama app, the default log file resides in the `$HOME/.ollama/logs/server.log`.
  - For the standalone binary, you should start it redirecting the logs to a file. E.g. `ollama serve > /tmp/ollama.log 2>&1`.
- `ConnectionString` returns the connection string to connect to the local Ollama binary started by the module instead of the container.
- `ContainerIP` returns the bound host IP `127.0.0.1` by default.
- `ContainerIPs` returns the bound host IP `["127.0.0.1"]` by default.
- `CopyToContainer`, `CopyDirToContainer`, `CopyFileToContainer` and `CopyFileFromContainer` return an error if called.
- `GetLogProductionErrorChannel` returns a nil channel.
- `Endpoint` returns the endpoint to connect to the local Ollama binary started by the module instead of the container.
- `Exec` passes the command to the local Ollama binary started by the module instead of inside the container. First argument is the command to execute, and the second argument is the list of arguments, else, an error is returned.
- `GetContainerID` returns the container ID of the local Ollama binary started by the module instead of the container, which maps to `local-ollama-<session-id>`.
- `Host` returns the bound host IP `127.0.0.1` by default.
- `Inspect` returns a ContainerJSON with the state of the local Ollama binary started by the module.
- `IsRunning` returns true if the local Ollama binary process started by the module is running.
- `Logs` returns the logs from the local Ollama binary started by the module instead of the container.
- `MappedPort` returns the port mapping for the local Ollama binary started by the module instead of the container.
- `Start` starts the local Ollama binary process.
- `State` returns the current state of the local Ollama binary process, `stopped` or `running`.
- `Stop` stops the local Ollama binary process.
- `Terminate` calls the `Stop` method and then removes the log file.

The local Ollama binary will create a log file in the current working directory, and it will be available in the container's `Logs` method.

!!!info
    The local Ollama binary will use the `OLLAMA_HOST` environment variable to set the host and port to listen on.
    If the environment variable is not set, it will default to `localhost:0`
    which bind to a loopback address on an ephemeral port to avoid port conflicts.

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
