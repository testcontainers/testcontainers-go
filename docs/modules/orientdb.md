# OrientDB

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [OrientDB](https://orientdb.org/), an open-source multi-model NoSQL database that combines graph, document, key/value, and object models in a single engine.

## Adding this module to your project dependencies

Please run the following command to add the OrientDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/orientdb
```

## Usage example

<!--codeinclude-->
[Creating a OrientDB container](../../modules/orientdb/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The OrientDB module exposes one entrypoint function to create the OrientDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "orientdb:3.2")`.

### Container Options

When starting the OrientDB container, you can pass options in a variadic way to configure it.

#### WithRootPassword

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the root password for the OrientDB instance by setting the `ORIENTDB_ROOT_PASSWORD` environment variable. The default password is `rootpwd`.

```golang
orientdb.WithRootPassword("mysecret")
```

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The OrientDB container exposes the following methods:

#### ServerURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `ServerURL` method returns the connection string for Java/JDBC clients using the OrientDB binary remote protocol, in the format `remote:<host>:<port>`.

```golang
serverURL, err := orientdbContainer.ServerURL(ctx)
// serverURL = "remote:localhost:2424"
```

#### StudioURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `StudioURL` method returns the URL for the OrientDB Studio web UI, in the format `http://<host>:<port>`.

```golang
studioURL, err := orientdbContainer.StudioURL(ctx)
// studioURL = "http://localhost:2480"
```
