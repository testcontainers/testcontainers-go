# Apache ActiveMQ Artemis

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Artemis.

## Adding this module to your project dependencies

Please run the following command to add the Artemis module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/artemis
```

## Usage example

<!--codeinclude-->
[Creating an Artemis container](../../modules/artemis/example_test.go) inside_block:runContainer
<!--/codeinclude-->

## Module reference

The Artemis module exposes one entrypoint function to create the Artemis container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ArtemisContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Artemis container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Artemis Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Artemis. E.g. `testcontainers.WithImage("docker.io/apache/activemq-artemis:2.30.0")`.

#### Wait Strategies

If you need to set a different wait strategy for Artemis, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for Artemis.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for Artemis, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

### Container Methods

The Artemis container exposes the following methods:



