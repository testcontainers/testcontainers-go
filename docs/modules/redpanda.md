# Redpanda

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

## Introduction

Redpanda is a streaming data platform for developers. Kafka API compatible. 10x faster. No ZooKeeper. No JVM!
This Testcontainers module provides three APIs:

- Kafka API
- Schema Registry API
- Redpanda Admin API

## Adding this module to your project dependencies

Please run the following command to add the Redpanda module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/redpanda
```

## Usage example

<!--codeinclude-->
[Creating a Redpanda container](../../modules/redpanda/examples_test.go) inside_block:runRedpandaContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Redpanda module exposes one entrypoint function to create the Redpanda container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*RedpandaContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Redpanda container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Redpanda Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "docker.redpanda.com/redpandadata/redpanda:v23.1.7")`.

{% include "../features/common_functional_options.md" %}

#### TLS Encryption

If you need to enable TLS use `WithTLS` with a valid PEM encoded certificate and key.

#### Additional Listener

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

There are scenarios where additional listeners are needed, for example if you
want to consume/from another container in the same network

You can use the `WithListener` option to add a listener to the Redpanda container.
<!--codeinclude-->
[Register additional listener](../../modules/redpanda/redpanda_test.go) inside_block:withListenerRP
<!--/codeinclude-->

Container defined in the same network
<!--codeinclude-->
[Start Kcat container](../../modules/redpanda/redpanda_test.go) inside_block:withListenerKcat
<!--/codeinclude-->

Produce messages using the new registered listener
<!--codeinclude-->
[Produce/consume via registered listener](../../modules/redpanda/redpanda_test.go) inside_block:withListenerExec
<!--/codeinclude-->

#### Adding Service Accounts

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

It's possible to add service accounts to the Redpanda container using the `WithNewServiceAccount` option, setting the service account name and its password.
E.g. `WithNewServiceAccount("service-account", "password")`.

#### Adding Super Users

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

When a super user is needed, you can use the `WithSuperusers` option, passing a variadic list of super users.
E.g. `WithSuperusers("superuser-1", "superuser-2")`.

#### Enabling SASL

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

The `WithEnableSASL()` option enables SASL scram sha authentication. By default, no authentication (plaintext) is used.
When setting an authentication method, make sure to add users as well and authorize them using the `WithSuperusers()` option.

#### WithEnableKafkaAuthorization

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

The `WithEnableKafkaAuthorization` enables authorization for connections on the Kafka API.

#### WithEnableWasmTransform

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

The `WithEnableWasmTransform` enables wasm transform.

!!!warning
    Should not be used with RP versions before 23.3

#### WithEnableSchemaRegistryHTTPBasicAuth

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

The `WithEnableSchemaRegistryHTTPBasicAuth` enables HTTP basic authentication for the Schema Registry.

#### WithAutoCreateTopics

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.22.0"><span class="tc-version">:material-tag: v0.22.0</span></a>

The `WithAutoCreateTopics` option enables the auto-creation of topics.

#### WithTLS

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

The `WithTLS` option enables TLS encryption. It requires a valid PEM encoded certificate and key, passed as byte slices.
E.g. `WithTLS([]byte(cert), []byte(key))`.

#### WithBootstrapConfig

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.33.0"><span class="tc-version">:material-tag: v0.33.0</span></a>

`WithBootstrapConfig` adds an arbitrary config key-value pair to the Redpanda container. Per the name, this config will be interpolated into the generated bootstrap
config file, which is particularly useful for configs requiring a restart when otherwise applied to a running Redpanda instance.
E.g. `WithBootstrapConfig("config_key", config_value)`, where `config_value` is of type `any`.

### Container Methods

The Redpanda container exposes the following methods:

#### KafkaSeedBroker

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

KafkaSeedBroker returns the seed broker that should be used for connecting
to the Kafka API with your Kafka client. It'll be returned in the format:
"host:port" - for example: "localhost:55687".

<!--codeinclude-->
[Get Kafka seed broker](../../modules/redpanda/redpanda_test.go) inside_block:kafkaSeedBroker
<!--/codeinclude-->

#### SchemaRegistryAddress

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

SchemaRegistryAddress returns the address to the schema registry API. This
is an HTTP-based API and thus the returned format will be: http://host:port.

<!--codeinclude-->
[Get schema registry address](../../modules/redpanda/redpanda_test.go) inside_block:schemaRegistryAddress
<!--/codeinclude-->


#### AdminAPIAddress

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

AdminAPIAddress returns the address to the Redpanda Admin API. This
is an HTTP-based API and thus the returned format will be: http://host:port.

<!--codeinclude-->
[Get admin API address](../../modules/redpanda/redpanda_test.go) inside_block:adminAPIAddress
<!--/codeinclude-->
