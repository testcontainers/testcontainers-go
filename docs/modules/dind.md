# DinD (Docker in Docker)

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

## Introduction

The Testcontainers module for DinD (Docker in Docker).

## Adding this module to your project dependencies

Please run the following command to add the DinD module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/dind
```

## Usage example

<!--codeinclude-->
[Creating a DinD container](../../modules/dind/examples_test.go) inside_block:runDinDContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

The DinD module exposes one entrypoint function to create the DinD container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DinDContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the DinD container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different DinD Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "docker:28.0.1-dind")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The DinD container exposes the following methods:

#### Host

The `Host` method returns the DinD URL, to be used for connecting
to the Docker API using a Docker client. It'll be returned in the format of `string`.

<!--codeinclude-->
[Host](../../modules/dind/examples_test.go) inside_block:didnHost
[Get a Docker client](../../modules/dind/examples_test.go) inside_block:getDockerClient
<!--/codeinclude-->

#### LoadImage

The `LoadImage` method loads an image into the docker in docker daemon.

This is useful for testing images generated locally without having to push them to a public docker registry.

The images must be already present in the node running the test. [DockerProvider](https://pkg.go.dev/github.com/testcontainers/testcontainers-go#DockerProvider) offers a method for pulling images, which can be used from the test code to ensure the image is present locally before loading them to the daemon.
