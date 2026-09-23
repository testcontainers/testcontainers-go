# Solr

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.44.0"><span class="tc-version">:material-tag: v0.44.0</span></a>

## Introduction

The Testcontainers module for Apache Solr, an open-source enterprise search platform built on Apache Lucene. It provides full-text search, faceting, hit highlighting, and real-time indexing capabilities.

## Adding this module to your project dependencies

Please run the following command to add the Solr module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/solr
```

## Usage example

<!--codeinclude-->
[Creating a Solr container](../../modules/solr/examples_test.go) inside_block:runSolrContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.44.0"><span class="tc-version">:material-tag: v0.44.0</span></a>

The Solr module exposes one entrypoint function to create the Solr container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "solr:9")`.

### Container Options

When starting the Solr container, you can pass options in a variadic way to configure it.

#### WithCollection

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.44.0"><span class="tc-version">:material-tag: v0.44.0</span></a>

`WithCollection` creates a named Solr collection after the container is ready.
The collection is created using the `solr create -c <name>` command and is
available immediately after `Run` returns.

```golang
solrContainer, err := solr.Run(ctx, "solr:9",
    solr.WithCollection("myCollection"),
)
```

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Solr container exposes the following methods:

#### Address

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.44.0"><span class="tc-version">:material-tag: v0.44.0</span></a>

`Address` returns the HTTP base URL of the Solr container in the form `http://<host>:<port>/solr`.

```golang
addr, err := solrContainer.Address(ctx)
```

#### CollectionURL

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.44.0"><span class="tc-version">:material-tag: v0.44.0</span></a>

`CollectionURL` returns the HTTP URL for a specific Solr collection in the form `http://<host>:<port>/solr/<collection>`.

```golang
url, err := solrContainer.CollectionURL(ctx, "myCollection")
```
