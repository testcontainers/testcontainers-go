# MongoDB

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for MongoDB.

## Adding this module to your project dependencies

Please run the following command to add the MongoDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mongodb
```

## Usage example

<!--codeinclude-->
[Creating a MongoDB container](../../modules/mongodb/mongodb_test.go) inside_block:createMongoDBContainer
<!--/codeinclude-->

## Module reference

The MongoDB module exposes one entrypoint function to create the MongoDB container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MongoDBContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the MongoDB container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different MongoDB Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for MongoDB. E.g. `testcontainers.WithImage("mongo:6")`.

#### Wait Strategies

If you need to set a different wait strategy for MongoDB, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for MongoDB.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for MongoDB, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

### Container Methods

The MongoDB container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the MongoDB container.
It returns a string with the format `mongodb://<host>:<port>`.

<!--codeinclude-->
[Get connection string](../../modules/mongodb/mongodb_test.go) inside_block:connectionString
<!--/codeinclude-->