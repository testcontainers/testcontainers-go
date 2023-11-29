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
[Creating a MongoDB container](../../modules/mongodb/mongodb_test.go) inside_block:runMongoDBContainer
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

#### WithUsername

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option sets the initial username to be created when the container starts.
It is used in conjunction with `WithPassword` to set a username and its password, creating the specified user with superuser power.

E.g. `testcontainers.WithUsername("mymongouser")`.

#### WithPassword

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option sets the initial password to be created when the container starts.
It is used in conjunction with `WithUsername` to set a username and its password, setting the password for the superuser power.

E.g. `testcontainers.WithPassword("mymongopwd")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The MongoDB container exposes the following methods:

#### ConnectionString

The `ConnectionString` method returns the connection string to connect to the MongoDB container.
It returns a string with the format `mongodb://<host>:<port>`.

It can be use to configure a MongoDB client (`go.mongodb.org/mongo-driver/mongo`), e.g.:

<!--codeinclude-->
[Using ConnectionString with the MongoDB client](../../modules/mongodb/mongodb_test.go) inside_block:connectToMongo
<!--/codeinclude-->
