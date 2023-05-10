# DockerRegistry

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Docker Registry.

## Adding this module to your project dependencies

Please run the following command to add the Docker Registry module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/dockerregistry
```

## Usage example

<!--codeinclude-->
[Creating a Docker Registry container](../../modules/dockerregistry/dockerregistry.go)
<!--/codeinclude-->

<!--codeinclude-->
[Test for a Docker Registry container](../../modules/dockerregistry/dockerregistry_test.go)
<!--/codeinclude-->

## Module reference

The Docker Registry module exposes one entrypoint function to create the Docker Registry container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DockerRegistryContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Docker Registry container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Docker Registry image, you can use `testcontainers.WithImage` with a valid Docker image
for DockerRegistry. E.g. `testcontainers.WithImage("docker.io/registry:latest")`.

#### Wait Strategies

If you need to set a different wait strategy for Docker Registry, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for DockerRegistry.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for Docker Registry, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

### Container Methods

The Docker Registry container exposes the following methods:
