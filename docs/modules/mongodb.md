# MongoDB

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.23.0"><span class="tc-version">:material-tag: v0.23.0</span></a>

## Introduction

The Testcontainers module for MongoDB.

## Adding this module to your project dependencies

Please run the following command to add the MongoDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mongodb
```

## Usage example

<!--codeinclude-->
[Creating a MongoDB container](../../modules/mongodb/examples_test.go) inside_block:runMongoDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The MongoDB module exposes one entrypoint function to create the MongoDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MongoDBContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the MongoDB container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "mongo:6")`.

#### WithUsername

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.27.0"><span class="tc-version">:material-tag: v0.27.0</span></a>

This functional option sets the initial username to be created when the container starts.
It is used in conjunction with `WithPassword` to set a username and its password, creating the specified user with superuser power.

E.g. `testcontainers.WithUsername("mymongouser")`.

#### WithPassword

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.27.0"><span class="tc-version">:material-tag: v0.27.0</span></a>

This functional option sets the initial password to be created when the container starts.
It is used in conjunction with `WithUsername` to set a username and its password, setting the password for the superuser power.

E.g. `testcontainers.WithPassword("mymongopwd")`.

#### WithReplicaSet

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.31.0"><span class="tc-version">:material-tag: v0.31.0</span></a>

The `WithReplicaSet` functional option configures the container to run a single-node MongoDB replica set named `rs`. The MongoDB container will wait until the replica set is ready.

{% include "../features/common_functional_options.md" %}

### Container Methods

The MongoDB container exposes the following methods:

#### ConnectionString

The `ConnectionString` method returns the connection string to connect to the MongoDB container.
It returns a string with the format `mongodb://<host>:<port>`.

It can be used to configure a MongoDB client (`go.mongodb.org/mongo-driver/mongo`), e.g.:

<!--codeinclude-->
[Using ConnectionString with the MongoDB client](../../modules/mongodb/examples_test.go) inside_block:connectToMongo
<!--/codeinclude-->
