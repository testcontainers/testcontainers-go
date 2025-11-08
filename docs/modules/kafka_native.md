# Kafka Native

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.39.0"><span class="tc-version">:material-tag: v0.39.0</span></a>

## Introduction

The Testcontainers module for [Apache Kafka Native](https://hub.docker.com/r/apache/kafka-native).

## Adding this module to your project dependencies

Please run the following command to add the Kafka module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/kafka_native
```

## Usage example

<!--codeinclude-->
[Creating a Kafka container](../../modules/kafka_native/examples_test.go) inside_block:runKafkaContainer
<!--/codeinclude-->

## Module Reference

### Run function

The Kafka module exposes one entrypoint function to create the Kafka container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "apache/kafka-native:3.9.1")`.

#### Environment variables

The environment variables that are already set by default are:

<!--codeinclude-->
[Environment variables](../../modules/kafka_native/kafka.go) inside_block:envVars
<!--/codeinclude-->

And also KAFKA_ADVERTISED_LISTENERS that is defined dynamically based on the container's hostname.

#### Init script

The Kafka container will be started using a custom shell script:

<!--codeinclude-->
[Init script](../../modules/kafka_native/kafka.go) inside_block:starterScript
<!--/codeinclude-->

### Container Options

When starting the Kafka container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Kafka container exposes the following methods:

#### Brokers

The `Brokers(ctx)` method returns the Kafka brokers as a string slice, containing the host and the random port defined by Kafka's public port (`9093/tcp`).

<!--codeinclude-->
[Get Kafka brokers](../../modules/kafka_native/kafka_test.go) inside_block:getBrokers
<!--/codeinclude-->
