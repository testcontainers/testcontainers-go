# Kafka

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

## Introduction

The Testcontainers module for Kafka.

This module runs Kafka in Kraft mode: [Apache Kafka Without ZooKeeper](https://developer.confluent.io/learn/kraft/) and it supports both [Apache Kafka](https://kafka.apache.org/) and [Confluent](https://docs.confluent.io/kafka/overview.html) images.

## Adding this module to your project dependencies

Please run the following command to add the Kafka module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/kafka
```

## Usage example

<!--codeinclude-->
[Apache Kafka Native](../../modules/kafka/examples_test.go) inside_block:runKafkaContainerApacheNative
<!--/codeinclude-->

<!--codeinclude-->
[Apache Kafka](../../modules/kafka/examples_test.go) inside_block:runKafkaContainerApacheNotNative
<!--/codeinclude-->

<!--codeinclude-->
[Confluent Kafka](../../modules/kafka/examples_test.go) inside_block:runKafkaContainerConfluentinc
<!--/codeinclude-->

The native container ([apache/kafka-native](https://hub.docker.com/r/apache/kafka-native/)) is based on GraalVM and typically starts several seconds faster than alternatives.

It is recommended to prefer Apache Kafka images over Confluent images, as Confluent has [unresolved issue with graceful shutdown](https://github.com/testcontainers/testcontainers-go/issues/2206).

Apache Kafka Native images are also smallest, however they do not include CLI tools such as `kafka-topics.sh`.

| Docker Image        | Size                     | Startup time    | Notes                   |
|---------------------|--------------------------|-----------------|-------------------------|
| Apache Kafka Native | 137MB (4.0.1 linux amd)  | ~1-3 seconds    | Does not have CLI tools |
| Apache Kafka        | 393MB (4.0.1 linux amd)  | ~4-5 seconds    |                         |
| Confluent Kafka     | 649MB (7.5.0 linux amd)  | ~13-15 seconds  | Shutdown issues         |

!!!info
    If you use image from custom registry, you might need to override starter script, see "Starter script" section below.

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

#### Starter script

The Kafka container will be started using a custom shell script.

Module would vary the starter script depending on the image in use, using following logic:

- image starts with `apache/kafka`: use Apache Kafka starter script.
- image starts with `confluentinc/`: use Confluent starter script.
- otherwise: use Confluent starter script (for backward compatibility).

This behavior can be overridden using the `kafka.WithApacheFlavor` or `kafka.WithConfluentFlavor` options. You can only provide one of these two options, otherwise an error would be returned when starting the container.

You can also provide a completely custom starter script using the `kafka.WithStarterScript` option, however note that if your script would become incompatible with the image in use, the container might fail to start.

<!--codeinclude-->
[Apache Kafka starter script](../../modules/kafka/kafka.go) inside_block:starterScriptApache
<!--/codeinclude-->

<!--codeinclude-->
[Confluent starter script](../../modules/kafka/kafka.go) inside_block:starterScriptConfluentinc
<!--/codeinclude-->

<!--codeinclude-->
[Overriding starter script](../../modules/kafka/examples_test.go) inside_block:runKafkaContainerWithApacheFlavor
<!--/codeinclude-->

### Container Options

When starting the Kafka container, you can pass options in a variadic way to configure it.

#### WithApacheFlavor/WithConfluentFlavor

You can manually specify which flavor of starter script to use with the following options:

<!--codeinclude-->
[With Apache Flavor](../../modules/kafka/examples_test.go) inside_block:runKafkaContainerWithApacheFlavor
<!--/codeinclude-->

<!--codeinclude-->
[With Confluent Flavor](../../modules/kafka/examples_test.go) inside_block:runKafkaContainerWithConfluentFlavor
<!--/codeinclude-->

#### WithStarterScript

This allows to provide a completely custom starter script for the Kafka container. Be careful when using this option, as compatibility with any image and module version cannot be guaranteed.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Kafka container exposes the following methods:

#### Brokers

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

The `Brokers(ctx)` method returns the Kafka brokers as a string slice, containing the host and the random port defined by Kafka's public port (`9093/tcp`).

<!--codeinclude-->
[Get Kafka brokers](../../modules/kafka/kafka_test.go) inside_block:getBrokers
<!--/codeinclude-->

## Localhost listener

Kafka container would by default be configured with `localhost:9095` as one of advertised listeners. This can be used when you need to run CLI commands inside the container, for example with custom wait strategies or to prepare test data.

Here is an example that uses custom wait strategy that checks if listing topics works:

<!--codeinclude-->
[Custom wait strategy](../../modules/kafka/examples_test.go) inside_block:runKafkaContainerAndUseLocalhostListener
<!--/codeinclude-->

Note: this will not work with `apache/kafka-native` images, as they do not include CLI tools.