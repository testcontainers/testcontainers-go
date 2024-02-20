# OpenSearch

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

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

## Module reference

The OpenSearch module exposes one entrypoint function to create the OpenSearch container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenSearchContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the OpenSearch container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different OpenSearch Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for OpenSearch. E.g. `testcontainers.WithImage("opensearchproject/opensearch:2.11.1")`.

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
