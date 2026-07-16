# RavenDB

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for RavenDB, an open-source NoSQL document database with multi-document ACID transactions. It is designed for web-scale, high-performance applications and provides a built-in Studio web UI for management and exploration.

## Adding this module to your project dependencies

Please run the following command to add the RavenDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/ravendb
```

## Usage example

<!--codeinclude-->
[Creating a RavenDB container](../../modules/ravendb/examples_test.go) inside_block:runRavenDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The RavenDB module exposes one entrypoint function to create the RavenDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "ravendb/ravendb:6.0-ubuntu-latest")`.

The module automatically sets the required environment variables for running RavenDB in unsecured development mode:

- `RAVEN_Setup_Mode=None`
- `RAVEN_License_Eula_Accepted=true`
- `RAVEN_Security_UnsecuredAccessAllowed=PublicNetwork`
- `RAVEN_Logs_Mode=None`

### Container Options

When starting the RavenDB container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The RavenDB container exposes the following methods:

#### ManagementURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `ManagementURL` method returns the URL of the RavenDB management interface (Studio UI and REST API).
It returns a string with the format `http://<host>:<port>`.

It can be used to configure a RavenDB client or to access the Studio UI in a browser, e.g.:

<!--codeinclude-->
[Using ManagementURL](../../modules/ravendb/examples_test.go) inside_block:ExampleRun_managementURL
<!--/codeinclude-->
