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

## Module reference

The Redpanda module exposes one entrypoint function to create the Redpanda container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RedpandaContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Redpanda container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Redpanda Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Redpanda. E.g. `testcontainers.WithImage("docker.redpanda.com/redpandadata/redpanda:v23.1.7")`.

{% include "../features/common_functional_options.md" %}

#### TLS Encryption

If you need to enable TLS use `WithTLS` with a valid PEM encoded certificate and key.

### Container Methods

The Redpanda container exposes the following methods:

#### KafkaSeedBroker

KafkaSeedBroker returns the seed broker that should be used for connecting
to the Kafka API with your Kafka client. It'll be returned in the format:
"host:port" - for example: "localhost:55687".

<!--codeinclude-->
[Get Kafka seed broker](../../modules/redpanda/redpanda_test.go) inside_block:kafkaSeedBroker
<!--/codeinclude-->

#### SchemaRegistryAddress

SchemaRegistryAddress returns the address to the schema registry API. This
is an HTTP-based API and thus the returned format will be: http://host:port.

<!--codeinclude-->
[Get schema registry address](../../modules/redpanda/redpanda_test.go) inside_block:schemaRegistryAddress
<!--/codeinclude-->


#### AdminAPIAddress

AdminAPIAddress returns the address to the Redpanda Admin API. This
is an HTTP-based API and thus the returned format will be: http://host:port.

<!--codeinclude-->
[Get admin API address](../../modules/redpanda/redpanda_test.go) inside_block:adminAPIAddress
<!--/codeinclude-->
