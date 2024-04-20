# DynamoDB

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

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

## Module reference

The DynamoDB module exposes one entrypoint function to create the DynamoDB container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DynamoDBContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the DynamoDB container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different DynamoDB Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for DynamoDB. E.g. `testcontainers.WithImage("amazon/dynamodb-local:2.4.0")`.

#### Tables

The DynamoDB module can initialize the container with one or more DynamoDB tables, via the `dynamodb.WithCreateTable` function.
This function accepts a `CreateTableInput` struct from the `github.com/aws/aws-sdk-go-v2/service/dynamodb` package, and can be used multiple times to create different tables.

{% include "../features/common_functional_options.md" %}

### Container Methods

The DynamoDB container exposes the following methods:

#### `DynamoDBClient`

The `DynamoDBClient` method creates an AWS SDK v2 DynamoDB client that can be used to interact with the DynamoDB container. This method accepts one parameter:

- `context.Context`, the Go context.
