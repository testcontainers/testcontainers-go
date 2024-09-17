# Kafka (KRaft)

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

## Introduction

The Testcontainers module for KRaft: [Apache Kafka Without ZooKeeper](https://developer.confluent.io/learn/kraft).

## Adding this module to your project dependencies

Please run the following command to add the Kafka module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/kafka
```

## Usage example

<!--codeinclude-->
[Creating a Kafka container](../../modules/kafka/examples_test.go) inside_block:runKafkaContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Kafka module exposes one entrypoint function to create the Kafka container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Kafka container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Kafka Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "confluentinc/confluent-local:7.5.0")`.

!!! warning
    The minimal required version of Kafka for KRaft mode is `confluentinc/confluent-local:7.4.0`. If you are using an image that
    is different from the official one, please make sure that it's compatible with KRaft mode, as the module won't check
    the version for you.

#### Init script

The Kafka container will be started using a custom shell script:

<!--codeinclude-->
[Init script](../../modules/kafka/kafka.go) inside_block:starterScript
<!--/codeinclude-->

#### Environment variables

The environment variables that are already set by default are:

<!--codeinclude-->
[Environment variables](../../modules/kafka/kafka.go) inside_block:envVars
<!--/codeinclude-->

{% include "../features/common_functional_options.md" %}


#### ClusterId

You can set up cluster id by using `WithClusterID` option.

```
KafkaContainer, err = kafka.Run(ctx,
		"confluentinc/confluent-local:7.6.1",
		kafka.WithClusterID("test-cluster"))
```

#### Listeners

If you need to connect new listeners, you can use `WithListener(listeners []KafkaListener)`. 
This option controls the following environment variables for the Kafka container: 
- `KAFKA_LISTENERS`
- `KAFKA_REST_BOOTSTRAP_SERVERS`
- `KAFKA_LISTENER_SECURITY_PROTOCOL_MAP`
- `KAFKA_INTER_BROKER_LISTENER_NAME`
- `KAFKA_ADVERTISED_LISTENERS`

Example:

<!--codeinclude-->
[Get Kafka brokers](../../modules/kafka/kafka_test.go) inside_block:kafkaWithListener
<!--/codeinclude-->

In the above code, we created a network for our container and attached kafka to it, so they can communicate. Then we marked port 9092 for our internal usage.

The first listener in the slice will be written in the env parameter `KAFKA_INTER_BROKER_LISTENER_NAME`  

Every listener's name will be converted in upper case. Every name and port should be unique and will be checked in a validation step.

If you are not using this option or the listeners list is empty, there will be 2 default listeners with the following addresses and ports:

External - Host():MappedPort()  
Internal - Host():9092

### Container Methods

The Kafka container exposes the following methods:

#### Brokers

The `Brokers(ctx)` method returns the Kafka brokers as a string slice, containing the host and the random port defined by Kafka's public port (`9093/tcp`).

<!--codeinclude-->
[Get Kafka brokers](../../modules/kafka/kafka_test.go) inside_block:getBrokers
<!--/codeinclude-->