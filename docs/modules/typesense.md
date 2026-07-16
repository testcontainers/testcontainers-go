# Typesense

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Typesense.

[Typesense](https://typesense.org/) is a blazing-fast, typo-tolerant, open-source search engine optimised for instant search-as-you-type experiences. It is an easy-to-use alternative to Algolia and a developer-friendly alternative to Elasticsearch.

## Adding this module to your project dependencies

Please run the following command to add the Typesense module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/typesense
```

## Usage example

<!--codeinclude-->
[Creating a Typesense container](../../modules/typesense/examples_test.go) inside_block:runTypesenseContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Typesense module exposes one entrypoint function to create the Typesense container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "typesense/typesense:26.0")`.

### Container Options

When starting the Typesense container, you can pass options in a variadic way to configure it.

#### WithAPIKey

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the API key used to authenticate requests to the Typesense container. The API key is required for all Typesense API calls. If not set, the default value `test-api-key` is used.

```golang
typesense.WithAPIKey("my-api-key")
```

#### WithDataDir

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the data directory inside the container where Typesense stores its data. If not set, the default value `/tmp` is used.

```golang
typesense.WithDataDir("/data")
```

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Typesense container exposes the following methods:

#### Address

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `Address` method retrieves the HTTP address of the Typesense container (e.g. `http://localhost:8108`).

#### APIKey

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `APIKey` method returns the API key configured for the Typesense container. The API key is required to authenticate all requests to the Typesense API.
