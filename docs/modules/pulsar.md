# Apache Pulsar

Testcontainers can be used to automatically create [Apache Pulsar](https://pulsar.apache.org) containers without external services.

It's based on the official Apache Pulsar docker image, so it is recommended to read the [official guide](https://pulsar.apache.org/docs/next/getting-started-docker/).

## Adding this module to your project dependencies

Please run the following command to add the LocalStack module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/pulsar
```

## Usage example

Create a `Pulsar` container to use it in your tests:

<!--codeinclude-->
[Creating a Pulsar container](../../modules/pulsar/pulsar_test.go) inside_block:startPulsarContainer
<!--/codeinclude-->

where the `tt.opts` are the options to configure the container. See the [Container Options](#container-options) section for more details.

Then you can retrieve the broker and the admin url:

<!--codeinclude-->
[Get broker and admin urls](../../modules/pulsar/pulsar_test.go) inside_block:getPulsarURLs
<!--/codeinclude-->

## Options

When starting the Pulsar container, you can pass options in a variadic way to configure it.

### Pulsar Configuration
If you need to set Pulsar configuration variables you can use the `WithPulsarEnv` to set Pulsar environment variables: the `PULSAR_PREFIX_` prefix will be automatically added for you.

For example, if you want to enable `brokerDeduplicationEnabled`:

<!--codeinclude-->
[Set configuration variables](../../modules/pulsar/pulsar_test.go) inside_block:addPulsarEnv
<!--/codeinclude-->

It will result in the `PULSAR_PREFIX_brokerDeduplicationEnabled=true` environment variable being set in the container request.

### Pulsar IO

If you need to test Pulsar IO framework you can enable the Pulsar Functions Worker with the `WithFunctionsWorker` option:

<!--codeinclude-->
[Create a Pulsar container with functions worker](../../modules/pulsar/pulsar_test.go) inside_block:withFunctionsWorker
<!--/codeinclude-->

### Pulsar Transactions

If you need to test Pulsar Transactions you can enable the transactions feature:

<!--codeinclude-->
[Create a Pulsar container with transactions](../../modules/pulsar/pulsar_test.go) inside_block:withTransactions
<!--/codeinclude-->
