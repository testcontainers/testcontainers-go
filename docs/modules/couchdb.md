# CouchDB

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Apache CouchDB, a document-oriented NoSQL database that uses JSON to store data, JavaScript as its query language using MapReduce, and HTTP for an API.

## Adding this module to your project dependencies

Please run the following command to add the CouchDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/couchdb
```

## Usage example

<!--codeinclude-->
[Creating a CouchDB container](../../modules/couchdb/examples_test.go) inside_block:runCouchDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The CouchDB module exposes one entrypoint function to create the CouchDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "couchdb:3")`.

### Container Options

When starting the CouchDB container, you can pass options in a variadic way to configure it.

#### WithAdminCredentials

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the admin username and password for the CouchDB container.
These credentials are used to authenticate against the CouchDB HTTP API and are reflected in the connection string returned by `ConnectionString`.

The default credentials are `admin` / `password`.

E.g. `couchdb.WithAdminCredentials("myuser", "mypassword")`.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The CouchDB container exposes the following methods:

#### ConnectionString

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `ConnectionString` method returns the connection string to connect to the CouchDB container.
It returns a string with the format `http://user:password@host:5984`.

E.g. `http://admin:password@localhost:32768`.

It can be used to configure a CouchDB HTTP client:

<!--codeinclude-->
[Using ConnectionString](../../modules/couchdb/examples_test.go) inside_block:runCouchDBContainer
<!--/codeinclude-->
