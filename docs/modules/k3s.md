# K3s

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.21.0"><span class="tc-version">:material-tag: v0.21.0</span></a>

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

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

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

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "rancher/k3s:v1.27.1-k3s1")`.

### Container Options

When starting the K3s container, you can pass options in a variadic way to configure it.

## WithManifest

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

The `WithManifest` option loads a manifest obtained from a local file into the cluster. K3s applies it automatically during the startup process

```golang
func WithManifest(manifestPath string) testcontainers.CustomizeRequestOption
```

Example:

```golang
        WithManifest("nginx-manifest.yaml")
```

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The K3s container exposes the following methods:

#### GetKubeConfig

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.21.0"><span class="tc-version">:material-tag: v0.21.0</span></a>

The `GetKubeConfig` method returns the K3s cluster's `kubeconfig`, including the server URL, to be used for connecting
to the Kubernetes Rest Client API using a Kubernetes client. It'll be returned in the format of `[]bytes`.

<!--codeinclude-->
[Get KubeConfig](../../modules/k3s/k3s_example_test.go) inside_block:GetKubeConfig
<!--/codeinclude-->

#### LoadImages

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.25.0"><span class="tc-version">:material-tag: v0.25.0</span></a>

The `LoadImages` method imports images from the local Docker daemon into the k3s cluster and makes them available to pods.

```golang
func (c *K3sContainer) LoadImages(ctx context.Context, images ...string) error
```

This is useful for testing images built locally without pushing them to a registry or configuring k3s to [use a private registry](https://docs.k3s.io/installation/private-registry).

Images must already be present on the Docker host running the test. [DockerProvider.PullImage](https://pkg.go.dev/github.com/testcontainers/testcontainers-go#DockerProvider.PullImage) is enough for single-architecture image references. When you need a specific OCI platform, use [DockerProvider.PullImageWithOpts](https://pkg.go.dev/github.com/testcontainers/testcontainers-go#DockerProvider.PullImageWithOpts) with [PullDockerImageWithPlatform](https://pkg.go.dev/github.com/testcontainers/testcontainers-go#PullDockerImageWithPlatform).

`LoadImages` delegates to `LoadImagesWithOpts` without save options. It works best with single-architecture image references (for example `amd64/nginx`). For multi-architecture images, use `LoadImagesWithPlatform` or `LoadImagesWithOpts`.

When creating pods that use loaded images, set `imagePullPolicy: Never` so Kubernetes uses the imported image instead of pulling from a registry.

Example:

```golang
provider, err := testcontainers.ProviderDocker.GetProvider()
if err != nil {
    // handle error
}

if err := provider.PullImage(ctx, "amd64/nginx"); err != nil {
    // handle error
}

if err := k3sContainer.LoadImages(ctx, "amd64/nginx"); err != nil {
    // handle error
}
```

#### LoadImagesWithPlatform

The `LoadImagesWithPlatform` method is a convenience wrapper around `LoadImagesWithOpts` that exports and imports an image for a specific OCI platform.

```golang
func (c *K3sContainer) LoadImagesWithPlatform(ctx context.Context, images []string, platform *v1.Platform) error
```

When `platform` is `nil`, behaviour matches `LoadImages`. When `platform` is set, the image is exported and imported for that OCI platform.

Use this method on multi-architecture hosts or when loading multi-architecture image tags. Pull the same platform into Docker first, then load it into k3s:

```golang
hostPlatform := platforms.DefaultSpec()
hostPlatform.OS = "linux"

provider, err := testcontainers.ProviderDocker.GetProvider()
if err != nil {
    // handle error
}

if err := provider.PullImageWithOpts(
    ctx,
    "amd64/nginx",
    testcontainers.PullDockerImageWithPlatform(hostPlatform),
); err != nil {
    // handle error
}

if err := k3sContainer.LoadImagesWithPlatform(ctx, []string{"amd64/nginx"}, &hostPlatform); err != nil {
    // handle error
}
```

#### LoadImagesWithOpts

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.25.0"><span class="tc-version">:material-tag: v0.25.0</span></a>

The `LoadImagesWithOpts` method imports local images into the k3s cluster, passing [SaveImageOption](https://pkg.go.dev/github.com/testcontainers/testcontainers-go#SaveImageOption) values through to `docker save`.

```golang
func (c *K3sContainer) LoadImagesWithOpts(ctx context.Context, images []string, opts ...testcontainers.SaveImageOption) error
```

When [SaveDockerImageWithPlatforms](https://pkg.go.dev/github.com/testcontainers/testcontainers-go#SaveDockerImageWithPlatforms) is passed, containerd import uses the same platform. Without platform options, import does not use `--all-platforms`.

For multiple images on different architectures, call `LoadImagesWithOpts` once per image and platform.

Example with platform:

```golang
hostPlatform := platforms.DefaultSpec()
hostPlatform.OS = "linux"

provider, err := testcontainers.ProviderDocker.GetProvider()
if err != nil {
    // handle error
}

if err := provider.PullImageWithOpts(
    ctx,
    "amd64/nginx",
    testcontainers.PullDockerImageWithPlatform(hostPlatform),
); err != nil {
    // handle error
}

if err := k3sContainer.LoadImagesWithOpts(
    ctx,
    []string{"amd64/nginx"},
    testcontainers.SaveDockerImageWithPlatforms(hostPlatform),
); err != nil {
    // handle error
}
```
