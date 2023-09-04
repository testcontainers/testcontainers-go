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
[Creating a Elasticsearch container](../../modules/elasticsearch/elasticsearch.go)
<!--/codeinclude-->

<!--codeinclude-->
[Test for a Elasticsearch container](../../modules/elasticsearch/elasticsearch_test.go)
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
for Elasticsearch. E.g. `testcontainers.WithImage("elasticsearch:8.0.0")`.

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

### Container Methods

The Elasticsearch container exposes the following methods:
