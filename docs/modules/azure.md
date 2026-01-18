# Azure

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

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
- [ServiceBus](#servicebus): `github.com/testcontainers/testcontainers-go/modules/azure/servicebus`.
!!! warning "EULA Acceptance"
    Due to licensing restrictions you are required to explicitly accept an End User License Agreement (EULA) for the EventHubs container image. This is facilitated through the `WithAcceptEULA` function.
- [CosmosDB](#cosmosdb): `github.com/testcontainers/testcontainers-go/modules/azure/cosmosdb`.
- [Lowkey Vault](#lowkey-vault): `github.com/testcontainers/testcontainers-go/modules/azure/lowkeyvault`.
<!--codeinclude-->
[Creating a Azurite container](../../modules/azure/azurite/examples_test.go) inside_block:runAzuriteContainer
<!--/codeinclude-->

## Azurite

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

The Azurite module exposes one entrypoint function to create the Azurite container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Default Credentials

The Azurite container uses the following default credentials:

<!--codeinclude-->
[Default Credentials](../../modules/azure/azurite/azurite.go) inside_block:defaultCredentials
<!--/codeinclude-->

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "mcr.microsoft.com/azure-storage/azurite:3.28.0")`.

### Container Options

When starting the Azurite container, you can pass options in a variadic way to configure it.

#### WithEnabledServices

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.40.0"><span class="tc-version">:material-tag: v0.40.0</span></a>

The default Azurite container entrypoint runs all three storage services: blob, queue, and table. Use this option to specify the required services for fewer exposed ports and slightly reduced container resources. E.g. `azurite.WithEnabledServices(azurite.BlobService)`.

#### WithInMemoryPersistence

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

If you want to use in-memory persistence, you can use `WithInMemoryPersistence(megabytes float64)`. E.g. `azurite.WithInMemoryPersistence(64.0)`.

Please read the [Azurite documentation](https://github.com/Azure/Azurite?tab=readme-ov-file#use-in-memory-storage) for more information.

!!! warning
    This option is only available in Azurite versions `3.28.0` and later.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Azurite container exposes the following methods:

#### BlobServiceURL

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

Returns the service URL to connect to the Blob service of the Azurite container and an error, passing the Go context as parameter.

#### QueueServiceURL

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

Returns the service URL to connect to the Queue service of the Azurite container and an error, passing the Go context as parameter.

#### TableServiceURL

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

Returns the service URL to connect to the Table service of the Azurite container and an error, passing the Go context as parameter.

### Examples

#### Blob Operations

In the following example, we will create a container with Azurite and perform some blob operations. For that, using the default
credentials, we will create an Azurite container, upload a blob to it, list the blobs, and download the blob. Finally, we will remove the created blob and container.

<!--codeinclude-->
[Create Container](../../modules/azure/azurite/examples_test.go) inside_block:runForBlobOperations
[Create Shared Key Credential](../../modules/azure/azurite/examples_test.go) inside_block:createSharedKeyCredential
[Create Client](../../modules/azure/azurite/examples_test.go) inside_block:createClient
[Create Container](../../modules/azure/azurite/examples_test.go) inside_block:createContainer
[Upload and Download Blob](../../modules/azure/azurite/examples_test.go) inside_block:uploadDownloadBlob
[List Blobs](../../modules/azure/azurite/examples_test.go) inside_block:listBlobs
[Delete Blob](../../modules/azure/azurite/examples_test.go) inside_block:deleteBlob
[Delete Container](../../modules/azure/azurite/examples_test.go) inside_block:deleteContainer
<!--/codeinclude-->

#### Queue Operations

In the following example, we will create an Azurite container and perform some queue operations. For that, using the default
credentials, we will create a queue, list the queues, and finally we will remove the created queue.

<!--codeinclude-->
[Run Azurite Container](../../modules/azure/azurite/examples_test.go) inside_block:runForQueueOperations
[Create Shared Key Credential](../../modules/azure/azurite/examples_test.go) inside_block:queueOperations_createSharedKeyCredential
[Create Client](../../modules/azure/azurite/examples_test.go) inside_block:queueOperations_createClient
[Create Queue](../../modules/azure/azurite/examples_test.go) inside_block:createQueue
[List Queues](../../modules/azure/azurite/examples_test.go) inside_block:listQueues
[Delete Queue](../../modules/azure/azurite/examples_test.go) inside_block:deleteQueue
<!--/codeinclude-->

#### Table Operations

In the following example, we will create an Azurite container and perform some table operations. For that, using the default
credentials, we will create a table, list the tables, and finally we will remove the created table.

<!--codeinclude-->
[Run Azurite Container](../../modules/azure/azurite/examples_test.go) inside_block:runForTableOperations
[Create Shared Key Credential](../../modules/azure/azurite/examples_test.go) inside_block:tableOperations_createSharedKeyCredential
[Create Client](../../modules/azure/azurite/examples_test.go) inside_block:tableOperations_createClient
[Create Table](../../modules/azure/azurite/examples_test.go) inside_block:createTable
[List Tables](../../modules/azure/azurite/examples_test.go) inside_block:listTables
[Delete Table](../../modules/azure/azurite/examples_test.go) inside_block:deleteTable
<!--/codeinclude-->

## EventHubs

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

The EventHubs module exposes one entrypoint function to create the EventHubs container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

The EventHubs container needs an Azurite container to be running, for that reason _Testcontainers for Go_ **automatically creates a Docker network and an Azurite container** for EventHubs to work.
When terminating the EventHubs container, the Azurite container and the Docker network are also terminated.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.0.1")`.

### Container Options

When starting the EventHubs container, you can pass options in a variadic way to configure it.

#### WithAzurite

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

This option allows you to set a different Azurite Docker image, instead of the default one, and also pass options to the Azurite container, in the form of a variadic argument of `testcontainers.ContainerCustomizer`.

#### WithAcceptEULA

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

This option allows you to accept the EULA for the EventHubs container.

#### WithConfig

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

This option allows you to set a custom EventHubs config file for the EventHubs container.

The config file must be a valid EventHubs config file, and it must be a valid JSON object.

<!--codeinclude-->
[EventHubs JSON Config](../../modules/azure/eventhubs/testdata/eventhubs_config.json)
<!--/codeinclude-->

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The EventHubs container exposes the following methods:

#### ConnectionString

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

Returns the connection string to connect to the EventHubs container and an error, passing the Go context as parameter.

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

## ServiceBus

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

The ServiceBus module exposes one entrypoint function to create the ServiceBus container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

The ServiceBus container needs a MSSQL Server container to be running, for that reason _Testcontainers for Go_ **automatically creates a Docker network and an MSSQL Server container** for ServiceBus to work.
When terminating the ServiceBus container, the MSSQL Server container and the Docker network are also terminated.

!!! info
    Since version `1.1.2` of the ServiceBus emulator, it's possible to set the `SQL_WAIT_INTERVAL` environment variable to the given seconds.
    This module sets it to `0` by default, because the MSSQL Server container is started first.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "mcr.microsoft.com/azure-messaging/servicebus-emulator:1.1.2")`.

### Container Options

When starting the ServiceBus container, you can pass options in a variadic way to configure it.

#### WithMSSQL

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

This option allows you to set a different MSSQL Server Docker image, instead of the default one, and also pass options to the MSSQL container, in the form of a variadic argument of `testcontainers.ContainerCustomizer`.

#### WithAcceptEULA

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

This option allows you to accept the EULA for the ServiceBus container.

#### WithConfig

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

This option allows you to set a custom ServiceBus config file for the ServiceBus container.

The config file must be a valid ServiceBus config file, and it must be a valid JSON object.

<!--codeinclude-->
[ServiceBus JSON Config](../../modules/azure/servicebus/testdata/servicebus_config.json)
<!--/codeinclude-->

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The ServiceBus container exposes the following methods:

#### ConnectionString

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.36.0"><span class="tc-version">:material-tag: v0.36.0</span></a>

Returns the connection string to connect to the ServiceBus container and an error, passing the Go context as parameter.

### Examples

#### Send events to ServiceBus

In the following example, inspired by the [Azure Event Hubs Go SDK](https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-go-get-started-send), we are creating an EventHubs container and sending events to it.

<!--codeinclude-->
[ServiceBus Config](../../modules/azure/servicebus/examples_test.go) inside_block:cfg
[Run ServiceBus Container](../../modules/azure/servicebus/examples_test.go) inside_block:runServiceBusContainer
[Create Client](../../modules/azure/servicebus/examples_test.go) inside_block:createClient
[Send messages to a Queue](../../modules/azure/servicebus/examples_test.go) inside_block:sendMessages
[Receive messages from a Queue](../../modules/azure/servicebus/examples_test.go) inside_block:receiveMessages
<!--/codeinclude-->

## CosmosDB

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.40.0"><span class="tc-version">:material-tag: v0.40.0</span></a>

The CosmosDB module exposes one entrypoint function to create the CosmosDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:vnext-preview")`.

### Container Options

When starting the CosmosDB container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The CosmosDB container exposes the following methods:

#### ConnectionString

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.40.0"><span class="tc-version">:material-tag: v0.40.0</span></a>

Returns the connection string to connect to the CosmosDB container and an error, passing the Go context as parameter.

### Examples

#### Connect and Create database

<!--codeinclude-->
[Connect_CreateDatabase](../../modules/azure/cosmosdb/examples_test.go) inside_block:ExampleRun_connect
<!--/codeinclude-->

## Lowkey Vault

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

The Lowkey Vault module exposes one entrypoint function to create the Lowkey Vault container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*LowkeyVaultContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.


#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "nagyesta/lowkey-vault:7.0.9-ubi10-minimal")`.

### Container Options

The Lowkey Vault container exposes two ports, one for the Key Vault API and one for the metadata endpoints such as the token endpoint.
Since the Key Vault API supports multiple vaults and selects the active vault based on the host authority of the request URL, the 
container can be configured in two ways:

#### Local mode

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

The default mode is to run the Key Vault container on localhost, meaning that both the Key Vault API and the metadata endpoints are 
exposed using random host ports, and the container is accessible only from the host machine. The default vault is automatically created 
and is made available using the host address. 

#### Network mode

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

To prepare the container for running in a network and making it accessible from other containers in the network, you can use the 
`WithNetworkAlias` option. For example:
```golang
Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal",
    lowkeyvault.WithNetworkAlias("lowkey-vault", aNetwork),
)
```

### Container Methods

The Lowkey Vault container exposes the following methods:

#### ConnectionUrl

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

Returns the connection URL to connect to the Key Vault API of the Lowkey Vault container and an error, passing the Go context and an 
`accessMode` as parameters. The access mode can be either `lowkeyvault.Local` or `lowkeyvault.Network` depending on the mode you wish
to use to connect to the Key Vault API.

#### TokenUrl

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

Returns the URL pointing to the token endpoint of the Lowkey Vault container and an error, passing the Go context and an `accessMode` 
as parameters. The access mode can be either `lowkeyvault.Local` or `lowkeyvault.Network` depending on the mode you wish
to use to access the token endpoint.

#### SetManagedIdentityEnvVariables

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

Can return an error, passing the Go context as the only parameter. This method conveniently sets the environment variables required 
to use managed identities with the Lowkey Vault container. When using the Azure Key Vault SDK for Go, you can authenticate with 
managed identities by using the `azidentity.NewDefaultAzureCredential(nil)` as a credential. In order for this authentication to work,
we need to set two environment variables, `IDENTITY_ENDPOINT` and `IDENTITY_HEADER` on the machine where the client code is running.
In case the client is running on the host, i.e., we are running the Lowkey Vault container in Local mode, this method can set both 
environment variables automatically.

#### PrepareClientForSelfSignedCert

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

Returns a `http.Client` and requires no parameters. This method can be used to prepare a `http.Client` for connecting to the Key Vault API 
of the Lowkey Vault container using a self-signed certificate. This is necessary since the Lowkey Vault container uses a self-signed 
certificate by default.

### Examples

#### Use the Secrets API in Local mode

In the following example, we are starting the Lowkey Vault container in Local mode, set a secret and retrieve it using the Key Vault Secrets API.

<!--codeinclude-->
[Run Lowkey Vault Container in Local mode](../../modules/azure/lowkeyvault/examples_test.go) inside_block:createContainer
[Create Client](../../modules/azure/lowkeyvault/examples_test.go) inside_block:prepareTheSecretClient
[Set and get a secret](../../modules/azure/lowkeyvault/examples_test.go) inside_block:setAndFetchTheSecret
<!--/codeinclude-->

#### Use the Keys API in Local mode

In the following example, we are starting the Lowkey Vault container in Local mode, create a key and encrypt and decrypt a message with it.

<!--codeinclude-->
[Run Lowkey Vault Container in Local mode](../../modules/azure/lowkeyvault/examples_test.go) inside_block:createContainer
[Create Client](../../modules/azure/lowkeyvault/examples_test.go) inside_block:prepareTheKeyClient
[Create a key](../../modules/azure/lowkeyvault/examples_test.go) inside_block:createKey
[Encrypt a message with the key](../../modules/azure/lowkeyvault/examples_test.go) inside_block:encryptMessage
[Decrypt cipher text with the key](../../modules/azure/lowkeyvault/examples_test.go) inside_block:decryptCipherText
<!--/codeinclude-->

#### Use the Certificates API in Local mode

In the following example, we are starting the Lowkey Vault container in Local mode, create a certificate using the Key Vault Certificates
API, and fetch the content of the certificate as a PKCS12 store using the Key Vault Secrets API.

<!--codeinclude-->
[Run Lowkey Vault Container in Local mode](../../modules/azure/lowkeyvault/examples_test.go) inside_block:createContainer
[Create Certificate Client](../../modules/azure/lowkeyvault/examples_test.go) inside_block:prepareTheCertClient
[Create a certificate](../../modules/azure/lowkeyvault/examples_test.go) inside_block:createCertificate
[Create Secret Client](../../modules/azure/lowkeyvault/examples_test.go) inside_block:prepareTheSecretClient
[Fetch the certificate details](../../modules/azure/lowkeyvault/examples_test.go) inside_block:fetchCertDetails
<!--/codeinclude-->

#### Use the Secrets API in Network mode

In the following example, we are starting the Lowkey Vault container in Network mode and setting the parameters of a Go client
container which will be used to connect to the Key Vault API of the Lowkey Vault container in Network mode.

<!--codeinclude-->
[Run Lowkey Vault Container in Network mode](../../modules/azure/lowkeyvault/examples_test.go) inside_block:createContainerWithNetwork
[Get the endpoint details in Network mode](../../modules/azure/lowkeyvault/examples_test.go) inside_block:obtainEndpointUrls
[Configure the client container](../../modules/azure/lowkeyvault/examples_test.go) inside_block:configureClient
<!--/codeinclude-->