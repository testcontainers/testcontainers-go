# Elasticsearch

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

## Introduction

The Testcontainers module for Elasticsearch.

## Adding this module to your project dependencies

Please run the following command to add the Elasticsearch module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/elasticsearch
```

## Usage example

<!--codeinclude-->
[Creating an Elasticsearch container](../../modules/elasticsearch/examples_test.go) inside_block:runElasticsearchContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Elasticsearch module exposes one entrypoint function to create the Elasticsearch container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ElasticsearchContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Elasticsearch container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "docker.elastic.co/elasticsearch/elasticsearch:8.0.0")`.

{% include "../features/common_functional_options.md" %}

#### Elasticsearch password

If you need to set a different password to request authorization when performing HTTP requests to the container, you can use the `WithPassword` option.  By default, the username is set to `elastic`, and the password is set to `changeme`.

!!!info
    In versions of Elasticsearch prior to 8.0.0, the default password is empty.

<!--codeinclude-->
[Custom Password](../../modules/elasticsearch/examples_test.go) inside_block:usingPassword
<!--/codeinclude-->

### Configuring the access to the Elasticsearch container

The Elasticsearch container exposes its settings in order to configure the client to connect to it. With those settings it's very easy to setup up our preferred way to connect to the container. We are going to show you two ways to connect to the container, using the HTTP client from the standard library, and using the Elasticsearch client.

!!!info
    The `TLS` access is only supported on Elasticsearch 8 and above, so please pay attention to how the below examples are using the `CACert` and `URL` settings.

#### Using the standard library's HTTP client

<!--codeinclude-->
[Create an HTTP client](../../modules/elasticsearch/elasticsearch_test.go) inside_block:createHTTPClient
<!--/codeinclude-->

The `esContainer` instance is obtained from the `elasticsearch.New` function.

In the case you configured the Elasticsearch container to set up a password, you'll need to add the `Authorization` header to the request. You can use the `SetBasicAuth` method from the HTTP request to generate the header value.

<!--codeinclude-->
[Using an authenticated client](../../modules/elasticsearch/elasticsearch_test.go) inside_block:basicAuthHeader
<!--/codeinclude-->

#### Using the Elasticsearch client

First, you must install the Elasticsearch Go client, so please read their [install guide](https://www.elastic.co/guide/en/elasticsearch/client/go-api/current/installation.html) for more information.

<!--codeinclude-->
[Create an Elasticsearch client](../../modules/elasticsearch/examples_test.go) inside_block:elasticsearchClient
<!--/codeinclude-->
