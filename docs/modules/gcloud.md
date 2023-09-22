# GCloud

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for GCloud.

## Adding this module to your project dependencies

Please run the following command to add the GCloud module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/gcloud
```

## Usage example

### BigTable

<!--codeinclude-->
[Creating a BigTable container](../../modules/gcloud/examples_test.go) inside_block:runBigTableContainer
<!--/codeinclude-->

### Datastore

<!--codeinclude-->
[Creating a Datastore container](../../modules/gcloud/examples_test.go) inside_block:runDatastoreContainer
<!--/codeinclude-->

## Module reference

The GCloud module exposes one entrypoint function to create the different GCloud emulators, and each function receives two parameters:

```golang
func RunBigTableContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*BigTableContainer, error)
func RunDatastoreContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DatastoreContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting any of the GCloud containers, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different GCloud Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for GCloud. E.g. `testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The GCloud container exposes the following methods:
