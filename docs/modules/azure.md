# Azure

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Azure.

## Adding this module to your project dependencies

Please run the following command to add the Azure module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/azure
```

## Usage example

The Azure module exposes the following Go packages:

- [Azurite](#azurite): `github.com/testcontainers/testcontainers-go/modules/azure/azurite`.
- [EventHubs](#eventhubs): `github.com/testcontainers/testcontainers-go/modules/azure/eventhubs`.

!!! warning "EULA Acceptance"
    Due to licensing restrictions you are required to explicitly accept an End User License Agreement (EULA) for the EventHubs container image. This is facilitated through the `WithAcceptEULA` function.

<!--codeinclude-->
[Creating a Azurite container](../../modules/azure/azurite/examples_test.go) inside_block:runAzuriteContainer
<!--/codeinclude-->

## Azurite

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Azurite module exposes one entrypoint function to create the Azurite container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*AzuriteContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Default Credentials

The Azurite container uses the following default credentials:

<!--codeinclude-->
[Default Credentials](../../modules/azure/azurite/azurite.go) inside_block:defaultCredentials
<!--/codeinclude-->

### Container Options

When starting the Azurite container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Azurite Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "mcr.microsoft.com/azure-storage/azurite:3.28.0")`.

{% include "../features/common_functional_options.md" %}

#### WithInMemoryPersistence

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you want to use in-memory persistence, you can use `WithInMemoryPersistence(megabytes float64)`. E.g. `azurite.WithInMemoryPersistence(64.0)`.

Please read the [Azurite documentation](https://github.com/Azure/Azurite?tab=readme-ov-file#use-in-memory-storage) for more information.

!!! warning
    This option is only available in Azurite versions 3.28.0 and later.

### Container Methods

The Azurite container exposes the following methods:

#### ServiceURL

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the service URL to connect to the Azurite container and an error, passing the Go context and the service name as parameters.

#### MustServiceURL

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the service URL to connect to the Azurite container, passing the Go context and the service name as parameters. If an error occurs, it will panic.

### Examples

#### Blob Operations

In the following example, we will create a container with Azurite and perform some blob operations. For that, using the default
credentials, we will create an Azurite container, upload a blob to it, list the blobs, and download the blob. Finally, we will remove the created blob and container.

<!--codeinclude-->
[Performing blob operations](../../modules/azure/azurite/examples_test.go) inside_block:blobOperations
<!--/codeinclude-->

#### Queue Operations

In the following example, we will create an Azurite container and perform some queue operations. For that, using the default
credentials, we will create a queue, list the queues, and finally we will remove the created queue.

<!--codeinclude-->
[Performing queue operations](../../modules/azure/azurite/examples_test.go) inside_block:queueOperations
<!--/codeinclude-->

#### Table Operations

In the following example, we will create an Azurite container and perform some table operations. For that, using the default
credentials, we will create a table, list the tables, and finally we will remove the created table.

<!--codeinclude-->
[Performing table operations](../../modules/azure/azurite/examples_test.go) inside_block:tableOperations
<!--/codeinclude-->

## EventHubs

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The EventHubs module exposes one entrypoint function to create the EventHubs container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

The EventHubs container needs an Azurite container to be running, for that reason _Testcontainers for Go_ automatically creates a Docker network and an Azurite container for EventHubs to work.
When terminating the EventHubs container, the Azurite container and the Docker network are also terminated.

### Container Options

When starting the Azurite container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Azurite Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "mcr.microsoft.com/azure-storage/azurite:3.28.0")`.

{% include "../features/common_functional_options.md" %}

#### WithAzuriteImage

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This option allows you to set a different Azurite Docker image, instead of the default one.

#### WithAcceptEULA

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This option allows you to accept the EULA for the EventHubs container.

#### WithConfig

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This option allows you to set a custom EventHubs config file for the EventHubs container.

The config file must be a valid EventHubs config file, and it must be a valid JSON object.

<!--codeinclude-->
[EventHubs JSON Config](../../modules/azure/eventhubs/testdata/eventhubs_config.json)
<!--/codeinclude-->

### Container Methods

The EventHubs container exposes the following methods:

#### ConnectionString

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the connection string to connect to the EventHubs container and an error, passing the Go context as parameter.

#### MustConnectionString

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the connection string to connect to the EventHubs container, passing the Go context as parameter. If an error occurs, it panics.

### Examples

#### Send events to EventHubs

In the following example, inspired by the [Azure Event Hubs Go SDK](https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-go-get-started-send), we are creating an EventHubs container and sending events to it.

<!--codeinclude-->
[EventHubs Config](../../modules/azure/eventhubs/examples_test.go) inside_block:cfg
[Run EventHubs Container](../../modules/azure/eventhubs/examples_test.go) inside_block:runEventHubsContainer
[Create Producer Client](../../modules/azure/eventhubs/examples_test.go) inside_block:createProducerClient
[Create Sample Events](../../modules/azure/eventhubs/examples_test.go) inside_block:createSampleEvents
[Create Batch](../../modules/azure/eventhubs/examples_test.go) inside_block:createBatch
[Send Event Data Batch to the EventHub](../../modules/azure/eventhubs/examples_test.go) inside_block:sendEventDataBatch
<!--/codeinclude-->