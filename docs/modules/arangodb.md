# ArangoDB

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for ArangoDB.

## Adding this module to your project dependencies

Please run the following command to add the ArangoDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/arangodb
```

## Usage example

<!--codeinclude-->
[Creating a ArangoDB container](../../modules/arangodb/examples_test.go) inside_block:runArangoDBContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The ArangoDB module exposes one entrypoint function to create the ArangoDB container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ArangoDBContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the ArangoDB container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "arangodb:3.11.5")`.

{% include "../features/common_functional_options.md" %}

#### WithRootPassword

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `WithRootPassword` function sets the root password for the ArangoDB container.

### Container Methods

The ArangoDB container exposes the following methods:

#### Credentials

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `Credentials` method returns the credentials for the ArangoDB container, in the form of a tuple of two strings: the username and the password.

```golang
func (c *Container) Credentials() (string, string)
```

#### HTTPEndpoint

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `HTTPEndpoint` method returns the HTTP endpoint of the ArangoDB container, using the following format: `http://$host:$port`.
