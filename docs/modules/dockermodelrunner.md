# Docker Model Runner

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

## Introduction

The Testcontainers module for DockerModelRunner.

## Adding this module to your project dependencies

Please run the following command to add the DockerModelRunner module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/dockermodelrunner
```

## Usage example

<!--codeinclude-->
[Creating a DockerModelRunner container](../../modules/dockermodelrunner/examples_test.go) inside_block:runWithModel
<!--/codeinclude-->

## Module Reference

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

The Docker Model Runner module exposes two entrypoint functions to create the Docker Model Runner container:

#### Run

This function receives two parameters:

```golang
func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

!!! info
    This function will use the default `socat` image under the hood. Please refer to the [socat module](../socat.md) for more information.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "alpine/socat:1.8.0.1")`.

### Container Options

When starting the Docker Model Runner container, you can pass options in a variadic way to configure it.

#### WithModel

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

Use the `WithModel` option to set the model to pull when the container is started. Please be aware, that only Models as OCI Artifacts are compatible with Docker Model Runner.

```golang
dockermodelrunner.WithModel("ai/llama3.2:latest")
```

!!! warning
    Multiple calls to this function overrides the previous value.

 You can find a curated collection of cutting-edge AI models as OCI Artifacts, from lightweight on-device models to high-performance LLMs on [Docker Hub](https://hub.docker.com/u/ai).

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Docker Model Runner container exposes the following methods:

#### PullModel

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

Use the `PullModel` method to pull a model from the Docker Model Runner. Make sure the passed context is not done before the pull operation is completed, so that the pull operation is cancelled.

<!--codeinclude-->
[Pulling a model at runtime](../../modules/dockermodelrunner/examples_test.go) inside_block:runPullModel
<!--/codeinclude-->

!!! info
     You can find a curated collection of cutting-edge AI models as OCI Artifacts, from lightweight on-device models to high-performance LLMs on [Docker Hub](https://hub.docker.com/u/ai).

#### InspectModel

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

Use the `InspectModel` method to inspect a model from the Docker Model Runner, by providing the model namespace and name.

<!--codeinclude-->
[Getting a model at runtime](../../modules/dockermodelrunner/examples_test.go) inside_block:runInspectModel
<!--/codeinclude-->

The namespace and name of the model is in the format of `<name>:<tag>`, which defines Models as OCI Artifacts in Docker Hub, therefore the namespace is the organization and the name is the repository.

E.g. `ai/smollm2:360M-Q4_K_M`. See [Models as OCI Artifacts](https://hub.docker.com/u/ai) for more information.

#### ListModels

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

Use the `ListModels` method to list all models that are already pulled locally, using the Docker Model Runner format.

<!--codeinclude-->
[Listing all models](../../modules/dockermodelrunner/examples_test.go) inside_block:runListModels
<!--/codeinclude-->
