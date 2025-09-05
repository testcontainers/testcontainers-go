# MongoDB Atlas Local

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The MongoDB Atlas Local module for Testcontainers lets you spin up a local MongoDB Atlas instance in Docker using
[mongodb/mongodb-atlas-local](https://hub.docker.com/r/mongodb/mongodb-atlas-local) for integration tests and
development. This module supports SCRAM authentication, init scripts, and custom log file mounting.

This module differs from the standard modules/mongodb Testcontainers module, allowing users to spin up a full local
Atlas-like environment complete with Atlas Search and Atlas Vector Search.

## Adding this module to your project dependencies

Please run the following command to add the MongoDB Atlas Local module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mongodb/atlaslocal
```

## Usage example

<!--codeinclude-->
[Creating a MongoDB Atlas Local container](../../modules/mongodb/atlaslocal/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `atlaslocal` module exposes one entrypoint function to create the MongoDB Atlas Local container, and this
function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "mongodb/mongodb-atlas-local:latest")`.

### Container Options

When starting the MongoDB Atlas Local container, you can pass options in a variadic way to configure it.

#### WithUsername

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option sets the initial username to be created when the container starts, populating the
`MONGODB_INITDB_ROOT_USERNAME` environment variable. You cannot mix this option with `WithUsernameFile`, as it will
result in an error.

#### WithPassword

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option sets the initial password to be created when the container starts, populating the
`MONGODB_INITDB_ROOT_PASSWORD` environment variable. You cannot mix this option with `WithPasswordFile`, as it will
result in an error.

#### WithUsernameFile

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option mounts a local file as the MongoDB root username secret at`/run/secrets/mongo-root-username`
and sets the `MONGODB_INITDB_ROOT_USERNAME_FILE` environment variable. The path must be absolute and exist; no-op if
empty.

#### WithPasswordFile

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option mounts a local file as the MongoDB root password secret at `/run/secrets/mongo-root-password` and
sets the `MONGODB_INITDB_ROOT_PASSWORD_FILE` environment variable. The path must be absolute and exist; no-op if empty.

#### WithNoTelemetry

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option disables the telemetry feature of MongoDB Atlas Local, setting the `DO_NOT_TRACK` environment
variable to `1`.

#### WithInitDatabase

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option allows you to specify a database name to be initialized when the container starts, populating
the `MONGODB_INITDB_DATABASE` environment variable.

#### WithInitScripts

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Mounts a directory into `/docker-entrypoint-initdb.d`, running `.sh`/`.js` scripts on startup. Calling this function
multiple times mounts only the latest directory.

#### WithMongotLogFile

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option writes the mongot logs to `/tmp/mongot.log` inside the container. See
`(*Container).ReadMongotLogs` to read the logs locally.

#### WithMongotLogToStdout

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option writes the mongot logs to `/dev/stdout` inside the container. See
`(*Container).ReadMongotLogs` to read the logs locally.

#### WithMongotLogToStderr

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option writes the mongot logs to `/dev/stderr` inside the container. See
`(*Container).ReadMongotLogs` to read the logs locally.

#### WithRunnerLogFile

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option writes the runner logs to `/tmp/runner.log` inside the container. See
`(*Container).ReadRunnerLogs` to read the logs locally.

#### WithRunnerLogToStdout

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option writes the runner logs to `/dev/stdout` inside the container. See
`(*Container).ReadRunnerLogs` to read the logs locally.

#### WithRunnerLogToStderr

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This functional option writes the runner logs to `/dev/stderr` inside the container. See
`(*Container).ReadRunnerLogs` to read the logs locally.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The MongoDB Atlas Local container exposes the following methods:


#### ConnectionString

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `ConnectionString` method returns the connection string to connect to the MongoDB Atlas Local container.
It returns a string with the format `mongodb://<host>:<port>[/<db>]/?directConnection=true[&authSource=admin]`.

It can be used to configure a MongoDB client (`go.mongodb.org/mongo-driver/v2/mongo`), e.g.:

<!--codeinclude-->
[Using ConnectionString with the MongoDB client](../../modules/mongodb/atlaslocal/examples_test.go) inside_block:connectToMongo
<!--/codeinclude-->

#### ReadMongotLogs

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `ReadMongotLogs` returns a reader for the log solution specified when constructing the container.


<!--codeinclude-->
[Using ReadMongotLogs with the MongoDB client](../../modules/mongodb/atlaslocal/examples_test.go) inside_block:readMongotLogs
<!--/codeinclude-->

#### ReadRunnerLogs

- Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `ReadRunnerLogs` returns a reader for the log solution specified when constructing the container.


<!--codeinclude-->
[Using ReadRunnerLogs with the MongoDB client](../../modules/mongodb/atlaslocal/examples_test.go) inside_block:readRunnerLogs
<!--/codeinclude-->
