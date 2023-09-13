# RabbitMQ

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for RabbitMQ.

## Adding this module to your project dependencies

Please run the following command to add the RabbitMQ module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/rabbitmq
```

## Usage example

<!--codeinclude-->
[Creating a RabbitMQ container](../../modules/rabbitmq/examples_test.go) inside_block:runRabbitMQContainer
<!--/codeinclude-->

## Module reference

The RabbitMQ module exposes one entrypoint function to create the RabbitMQ container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RabbitMQContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the RabbitMQ container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different RabbitMQ Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for RabbitMQ. E.g. `testcontainers.WithImage("rabbitmq:3.12-management-alpine")`.

#### Wait Strategies

If you need to set a different wait strategy for RabbitMQ, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for RabbitMQ.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for RabbitMQ, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

### Container Methods

The RabbitMQ container exposes the following methods:
