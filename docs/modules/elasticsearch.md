# Elasticsearch

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Elasticsearch.

## Adding this module to your project dependencies

Please run the following command to add the Elasticsearch module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/elasticsearch
```

## Usage example

<!--codeinclude-->
[Creating a Elasticsearch container](../../modules/elasticsearch/examples_test.go) inside_block:runElasticsearchContainer
<!--/codeinclude-->

## Module reference

The Elasticsearch module exposes one entrypoint function to create the Elasticsearch container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ElasticsearchContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Elasticsearch container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Elasticsearch Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Elasticsearch. E.g. `testcontainers.WithImage("docker.elastic.co/elasticsearch/elasticsearch:8.0.0")`.

#### Wait Strategies

If you need to set a different wait strategy for Elasticsearch, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for Elasticsearch.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for Elasticsearch, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

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

The `esContainer` instance is obtained from the `elasticsearch.RunContainer` function.

In the case you configured the Elasticsearch container to set up a password, you'll need to add the `Authorization` header to the request. You can use the `SetBasicAuth` method from the HTTP request to generate the header value.

<!--codeinclude-->
[Using an authenticated client](../../modules/elasticsearch/elasticsearch_test.go) inside_block:basicAuthHeader
<!--/codeinclude-->

#### Using the Elasticsearch client

First, you must install the Elasticsearch Go client, so please read their [install guide](https://www.elastic.co/guide/en/elasticsearch/client/go-api/current/installation.html) for more information.

<!--codeinclude-->
[Create an Elasticsearch client](../../modules/elasticsearch/examples_test.go) inside_block:elasticsearchClient
<!--/codeinclude-->
