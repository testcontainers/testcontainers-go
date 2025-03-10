# K3s

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.21.0"><span class="tc-version">:material-tag: v0.21.0</span></a>

## Introduction

The Testcontainers module for K3s.

## Adding this module to your project dependencies

Please run the following command to add the K3s module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/k3s
```

## Usage example

<!--codeinclude-->
[Test for a K3s container](../../modules/k3s/k3s_test.go) inside_block:runK3sContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The K3s module exposes one entrypoint function to create the K3s container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*K3sContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Ports
These are the ports used by the K3s container:
<!--codeinclude-->
[Container Ports](../../modules/k3s/k3s.go) inside_block:containerPorts
<!--/codeinclude-->

### Container Options

When starting the K3s container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "rancher/k3s:v1.27.1-k3s1")`.

{% include "../features/common_functional_options.md" %}

## WithManifest

The `WithManifest` option loads a manifest obtained from a local file into the cluster. K3s applies it automatically during the startup process

```golang
func WithManifest(manifestPath string) testcontainers.CustomizeRequestOption
```

Example:

```golang
        WithManifest("nginx-manifest.yaml")
```

### Container Methods

The K3s container exposes the following methods:

#### GetKubeConfig

The `GetKubeConfig` method returns the K3s cluster's `kubeconfig`, including the server URL, to be used for connecting
to the Kubernetes Rest Client API using a Kubernetes client. It'll be returned in the format of `[]bytes`.

<!--codeinclude-->
[Get KubeConfig](../../modules/k3s/k3s_example_test.go) inside_block:GetKubeConfig
<!--/codeinclude-->

#### LoadImages

The `LoadImages` method loads a list of images into the kubernetes cluster and makes them available to pods.

This is useful for testing images generated locally without having to push them to a public docker registry or having to configure `k3s` to [use a private registry](https://docs.k3s.io/installation/private-registry).

The images must be already present in the node running the test. [DockerProvider](https://pkg.go.dev/github.com/testcontainers/testcontainers-go#DockerProvider) offers a method for pulling images, which can be used from the test code to ensure the image is present locally before loading them to the cluster.
