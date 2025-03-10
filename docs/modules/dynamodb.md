# DynamoDB

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

## Introduction

The Testcontainers module for DynamoDB.

## Adding this module to your project dependencies

Please run the following command to add the DynamoDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/dynamodb
```

## Usage example

<!--codeinclude-->
[Creating a DynamoDB container](../../modules/dynamodb/examples_test.go) inside_block:runDynamoDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

The DynamoDB module exposes one entrypoint function to create the DynamoDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DynamoDBContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the DynamoDB container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "amazon/dynamodb-local:2.2.1")`.

{% include "../features/common_functional_options.md" %}

#### WithSharedDB

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

The `WithSharedDB` option tells the DynamoDB container to use a single database file. At the same time, it marks the container as reusable, which causes that successive calls to the `Run` function will return the same container instance, and therefore, the same database file.

#### WithDisableTelemetry

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

You can turn off telemetry when starting the DynamoDB container, using the option `WithDisableTelemetry`.

### Container Methods

The DynamoDB container exposes the following methods:

#### ConnectionString

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

The `ConnectionString` method returns the connection string to the DynamoDB container. This connection string can be used to connect to the DynamoDB container from your application,
using the AWS SDK or any other DynamoDB client of your choice.

<!--codeinclude-->
[Creating a client](../../modules/dynamodb/dynamodb_test.go) inside_block:createClient
<!--/codeinclude-->

The above example uses `github.com/aws/aws-sdk-go-v2/service/dynamodb` to create a client and connect to the DynamoDB container.
