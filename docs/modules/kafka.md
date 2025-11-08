# Kafka

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

## Introduction

The Testcontainers module for Kafka.

This module would run Kafka in Kraft mode: [Apache Kafka Without ZooKeeper](https://developer.confluent.io/learn/kraft/) and it supports both [Apache Kafka](https://kafka.apache.org/) and [Confluent](https://docs.confluent.io/kafka/overview.html) images.

## Adding this module to your project dependencies

Please run the following command to add the Kafka module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/kafka
```

## Usage example

<!--codeinclude-->
[Apache Native Kafka](../../modules/kafka/examples_test.go) inside_block:runKafkaContainerApacheNative
<!--/codeinclude-->

<!--codeinclude-->
[Apache Kafka](../../modules/kafka/examples_test.go) inside_block:runKafkaContainerApacheNotNative
<!--/codeinclude-->

<!--codeinclude-->
[Confluent Kafka](../../modules/kafka/examples_test.go) inside_block:runKafkaContainerConfluent
<!--/codeinclude-->

The native container ([apache/kafka-native](https://hub.docker.com/r/apache/kafka-native/)) is based on GraalVM and typically starts several seconds faster than alternatives.

## Module Reference

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Kafka module exposes one entrypoint function to create the Kafka container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "apache/kafka-native:4.0.1")`.

!!! warning
    Module expects that the image in use supports Kraft mode (Kafka without ZooKeeper).
    The minimal required version of Confluent images for KRaft mode is `confluentinc/confluent-local:7.4.0`.
    All Apache images support Kraft mode.

#### Environment variables

The environment variables that are already set by default are:

<!--codeinclude-->
[Environment variables](../../modules/kafka/kafka.go) inside_block:envVars
<!--/codeinclude-->

#### Init script

The Kafka container will be started using a custom shell script.

Module would vary the starter script depending on the image in use, using following logic:

- image starts with `apache/kafka`: use Apache Kafka starter script.
- image starts with `confluentinc/`: use Confluent starter script.

<!--codeinclude-->
[Apache Kafka starter script](../../modules/kafka/kafka.go) inside_block:starterScriptApache
<!--/codeinclude-->

<!--codeinclude-->
[Confluent starter script](../../modules/kafka/kafka.go) inside_block:starterScriptConfluentinc
<!--/codeinclude-->

### Container Options

When starting the Kafka container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Kafka container exposes the following methods:

#### Brokers

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

The `Brokers(ctx)` method returns the Kafka brokers as a string slice, containing the host and the random port defined by Kafka's public port (`9093/tcp`).

<!--codeinclude-->
[Get Kafka brokers](../../modules/kafka/kafka_test.go) inside_block:getBrokers
<!--/codeinclude-->
