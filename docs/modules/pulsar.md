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
[Creating a Pulsar container](../../modules/pulsar/pulsar_test.go) inside_block:startPulsarContainer
<!--/codeinclude-->

where the `tt.opts` are the options to configure the container. See the [Container Options](#container-options) section for more details.

## Module Reference

The Redis module exposes one entrypoint function to create the containerr, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Pulsar container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Pulsar Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Pulsar. E.g. `testcontainers.WithImage("docker.io/apachepulsar/pulsar:2.10.2")`.

<!--codeinclude-->
[Set Pulsar image](../../modules/pulsar/pulsar_test.go) inside_block:setPulsarImage
<!--/codeinclude-->

#### Wait Strategies

If you need to set a different wait strategy for Pulsar, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for Pulsar.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for Pulsar, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

<!--codeinclude-->
[Advanced Docker settings](../../modules/pulsar/pulsar_test.go) inside_block:advancedDockerSettings
<!--/codeinclude-->

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

#### Log consumers
If you need to collect the logs from the Pulsar container, you can add your own LogConsumer with the `WithLogConsumers` function, which accepts a variadic argument of LogConsumers.

<!--codeinclude-->
[Adding LogConsumers](../../modules/pulsar/pulsar_test.go) inside_block:withLogConsumers
<!--/codeinclude-->

An example of a LogConsumer could be the following:

<!--codeinclude-->
[Example LogConsumer](../../modules/pulsar/pulsar_test.go) inside_block:logConsumerForTesting
<!--/codeinclude-->

!!!warning
    You will need to explicitly stop the producer in your tests.

If you want to know more about LogConsumers, please check the [Following Container Logs](../features/follow_logs.md) documentation.

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
