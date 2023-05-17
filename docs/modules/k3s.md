# K3s

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for K3s.

## Adding this module to your project dependencies

Please run the following command to add the K3s module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/k3s
```

## Usage example

<!--codeinclude-->
[Creating a K3s container](../../modules/k3s/k3s.go)
<!--/codeinclude-->

<!--codeinclude-->
[Test for a K3s container](../../modules/k3s/k3s_test.go)
<!--/codeinclude-->

## Module reference

The K3s module exposes one entrypoint function to create the K3s container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*K3sContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the K3s container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different K3s Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for K3s. E.g. `testcontainers.WithImage("docker.io/rancher/k3s:latest")`.

#### Wait Strategies

If you need to set a different wait strategy for K3s, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for K3s.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for K3s, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

### Container Methods

The K3s container exposes the following methods:
