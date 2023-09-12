# Kafka

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Kafka.

## Adding this module to your project dependencies

Please run the following command to add the Kafka module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/kafka
```

## Usage example

<!--codeinclude-->
[Creating a Kafka container](../../modules/kafka/examples_test.go) inside_block:runKafkaContainer
<!--/codeinclude-->

## Module reference

The Kafka module exposes one entrypoint function to create the Kafka container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Kafka container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Kafka Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Kafka. E.g. `testcontainers.WithImage("confluentinc/cp-kafka:7.3.3")`.

#### Wait Strategies

If you need to set a different wait strategy for Kafka, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for Kafka.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for Kafka, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

### Container Methods

The Kafka container exposes the following methods:
