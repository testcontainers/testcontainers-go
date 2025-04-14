# OpenSearch

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

## Introduction

The Testcontainers module for OpenSearch.

## Adding this module to your project dependencies

Please run the following command to add the OpenSearch module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/opensearch
```

## Usage example

<!--codeinclude-->
[Creating a OpenSearch container](../../modules/opensearch/examples_test.go) inside_block:runOpenSearchContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The OpenSearch module exposes one entrypoint function to create the OpenSearch container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OpenSearchContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the OpenSearch container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "opensearchproject/opensearch:2.11.1")`.

{% include "../features/common_functional_options.md" %}

#### User and password

If you need to set a different password to request authorization when performing HTTP requests to the container, you can use the `WithUsername` and `WithPassword` options. By default, the username is set to `admin`, and the password is set to `admin`.

<!--codeinclude-->
[Custom Credentials](../../modules/opensearch/examples_test.go) inside_block:runOpenSearchContainer
<!--/codeinclude-->

### Container Methods

The OpenSearch container exposes the following methods:

#### Address

The `Address` method returns the location where the OpenSearch container is listening.
It returns a string with the format `http://<host>:<port>`.

!!!warning
    TLS is not supported at the moment.

<!--codeinclude-->
[Connecting using HTTP](../../modules/opensearch/opensearch_test.go) inside_block:httpConnection
<!--/codeinclude-->
