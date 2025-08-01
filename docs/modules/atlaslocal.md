# MongoDBAtlasLocal

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for MongoDBAtlasLocal.

## Adding this module to your project dependencies

Please run the following command to add the MongoDBAtlasLocal module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/atlaslocal
```

## Usage example

<!--codeinclude-->
[Creating a MongoDBAtlasLocal container](../../modules/atlaslocal/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The MongoDBAtlasLocal module exposes one entrypoint function to create the MongoDBAtlasLocal container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "mongodb/mongodb-atlas-local:latest")`.

### Container Options

When starting the MongoDBAtlasLocal container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The MongoDBAtlasLocal container exposes the following methods:
