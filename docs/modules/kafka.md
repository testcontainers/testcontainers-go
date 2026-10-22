# Kafka (KRaft)

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

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
In example: `Run(context.Background(), "confluentinc/confluent-local:7.5.0")`.

!!! warning
    The minimal required version of Kafka for KRaft mode is `confluentinc/confluent-local:7.4.0`. If you are using an image that
    is different from the official one, please make sure that it's compatible with KRaft mode, as the module won't check
    the version for you.

#### Environment variables

The environment variables that are already set by default are:

<!--codeinclude-->
[Environment variables](../../modules/kafka/kafka.go) inside_block:envVars
<!--/codeinclude-->

#### Init script

The Kafka container will be started using a custom shell script:

<!--codeinclude-->
[Init script](../../modules/kafka/kafka.go) inside_block:starterScript
<!--/codeinclude-->

### Container Options

When starting the Kafka container, you can pass options in a variadic way to configure it.

#### WithTLS

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you need to test clients against a TLS-secured broker, you can use the `kafka.WithTLS()` option.

When enabled, the container will:

- Generate a self-signed CA and a server certificate signed by it, valid for the host the container is reachable at
- Configure Kafka with a PKCS12 keystore holding the server certificate
- Expose an additional SSL listener on port `9095`, keeping the PLAINTEXT one on port `9093`

If you override `KAFKA_LISTENERS` or `KAFKA_LISTENER_SECURITY_PROTOCOL_MAP`, the SSL listener is still added. If you define it yourself, it must listen on port `9095` and use the `SSL` protocol, otherwise `Run` returns an error.

Use the `BrokersTLS(ctx)` method to get the SSL endpoint and the `TLSConfig()` method to get the `*tls.Config` for client connections.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Kafka container exposes the following methods:

#### Brokers

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

The `Brokers(ctx)` method returns the Kafka brokers as a string slice, containing the host and the random port defined by Kafka's public port (`9093/tcp`).

<!--codeinclude-->
[Get Kafka brokers](../../modules/kafka/kafka_test.go) inside_block:getBrokers
<!--/codeinclude-->

#### BrokersTLS

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `BrokersTLS(ctx)` method returns the Kafka brokers of the SSL listener as a string slice, containing the host and the random port defined by Kafka's SSL port (`9095/tcp`). It can only be used when the container was created with the `WithTLS()` option. Returns an error if TLS was not enabled.

#### TLSConfig

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `TLSConfig()` method returns the TLS configuration trusting the CA that signed the broker certificate. It can only be used when the container was created with the `WithTLS()` option. Returns an error if TLS was not enabled.

<!--codeinclude-->
[Get Kafka TLS brokers and config](../../modules/kafka/kafka_test.go) inside_block:getBrokersTLS
<!--/codeinclude-->
