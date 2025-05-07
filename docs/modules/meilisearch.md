# Meilisearch

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

## Introduction

The Testcontainers module for Meilisearch.

## Adding this module to your project dependencies

Please run the following command to add the Meilisearch module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/meilisearch
```

## Usage example

<!--codeinclude-->
[Creating a Meilisearch container](../../modules/meilisearch/examples_test.go) inside_block:runMeilisearchContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

The Meilisearch module exposes one entrypoint function to create the Meilisearch container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MeilisearchContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Meilisearch container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "getmeili/meilisearch:v1.10.3")`.

{% include "../features/common_functional_options.md" %}

#### Master Key

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

If you need to set a master key, you can use the `WithMasterKey(key string)` option. Otherwise, the default will be used which is `just-a-master-key-for-test`, which is exported on the container fields.

#### Dump Data

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

If you need to dump data in Meilisearch upon initialization for testing, you can use `WithDumpData(filepath string)` option where `filepath` can be an absolute path or relative path to a `dump` file. Please refer to the official Meilisearch documentation about dump files [here](https://www.meilisearch.com/docs/learn/advanced/snapshots_vs_dumps#dumps).

### Container Methods

The Meilisearch container exposes the following methods:

#### Address

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

The `Address` method retrieves the address of the Meilisearch container.
It will use http as protocol, as TLS is not supported at the moment.

#### MasterKey

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.34.0"><span class="tc-version">:material-tag: v0.34.0</span></a>

The `MasterKey` method retrieves the master key of the Meilisearch container.
