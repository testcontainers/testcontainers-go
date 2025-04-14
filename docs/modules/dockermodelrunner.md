# Docker Model Runner

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for DockerModelRunner.

## Adding this module to your project dependencies

Please run the following command to add the DockerModelRunner module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/dockermodelrunner
```

## Usage example

<!--codeinclude-->
[Creating a DockerModelRunner container](../../modules/dockermodelrunner/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Docker Model Runner module exposes one entrypoint function to create the Docker Model Runner container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Docker Model Runner container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "alpine/socat:1.8.0.1")`.

{% include "../features/common_functional_options.md" %}

#### WithModel

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use the `WithModel` option to set the model to pull when the container is started.

```golang
dockermodelrunner.WithModel("ai/llama3.2:latest")
```

!!! warning
    Multiple calls to this function overrides the previous value.

 You can find a curated collection of cutting-edge AI models as OCI Artifacts, from lightweight on-device models to high-performance LLMs on Docker Hub: https://hub.docker.com/u/ai.

### Container Methods

The Docker Model Runner container exposes the following methods:

#### PullModel

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use the `PullModel` method to pull a model from the Docker Model Runner.

!!! info
     You can find a curated collection of cutting-edge AI models as OCI Artifacts, from lightweight on-device models to high-performance LLMs on Docker Hub: https://hub.docker.com/u/ai.

#### ListModels

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use the `ListModels` method to list all models that are already pulled locally, using the Docker Model Runner format.
