# Apache Pulsar

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.19.0"><span class="tc-version">:material-tag: v0.19.0</span></a>

## Introduction

The Testcontainers module for Apache Pulsar.

Testcontainers can be used to automatically create [Apache Pulsar](https://pulsar.apache.org) containers without external services.

It's based on the official Apache Pulsar docker image, so it is recommended to read the [official guide](https://pulsar.apache.org/docs/next/getting-started-docker/).

## Adding this module to your project dependencies

Please run the following command to add the Apache Pulsar module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/pulsar
```

## Usage example

Create a `Pulsar` container to use it in your tests:

<!--codeinclude-->
[Creating a Pulsar container](../../modules/pulsar/examples_test.go) inside_block:runPulsarContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Pulsar module exposes one entrypoint function to create the container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Pulsar container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "apachepulsar/pulsar:2.10.2")`.

{% include "../features/common_functional_options.md" %}

<!--codeinclude-->
[Advanced Docker settings](../../modules/pulsar/pulsar_test.go) inside_block:advancedDockerSettings
<!--/codeinclude-->

Here, the `nwName` relates to the name of a previously created Docker network. Please see the [How to create a network](../features/creating_networks.md) documentation for more information.

#### Pulsar Configuration
If you need to set Pulsar configuration variables you can use the `WithPulsarEnv` to set Pulsar environment variables: the `PULSAR_PREFIX_` prefix will be automatically added for you.

For example, if you want to enable `brokerDeduplicationEnabled`:

<!--codeinclude-->
[Set configuration variables](../../modules/pulsar/pulsar_test.go) inside_block:addPulsarEnv
<!--/codeinclude-->

It will result in the `PULSAR_PREFIX_brokerDeduplicationEnabled=true` environment variable being set in the container request.

#### Pulsar IO

If you need to test Pulsar IO framework you can enable the Pulsar Functions Worker with the `WithFunctionsWorker` option:

<!--codeinclude-->
[Create a Pulsar container with functions worker](../../modules/pulsar/pulsar_test.go) inside_block:withFunctionsWorker
<!--/codeinclude-->

#### Pulsar Transactions

If you need to test Pulsar Transactions you can enable the transactions feature:

<!--codeinclude-->
[Create a Pulsar container with transactions](../../modules/pulsar/pulsar_test.go) inside_block:withTransactions
<!--/codeinclude-->

### Container methods

Once you have a Pulsar container, then you can retrieve the broker and the admin url:

#### Admin URL

<!--codeinclude-->
[Get admin url](../../modules/pulsar/pulsar_test.go) inside_block:getAdminURL
<!--/codeinclude-->

#### Broker URL

<!--codeinclude-->
[Get broker url](../../modules/pulsar/pulsar_test.go) inside_block:getBrokerURL
<!--/codeinclude-->
