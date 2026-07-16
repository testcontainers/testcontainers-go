# FakeGCSServer

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [FakeGCSServer](https://github.com/fsouza/fake-gcs-server) provides a Google Cloud Storage API emulator for local development and testing.
It runs [fsouza/fake-gcs-server](https://github.com/fsouza/fake-gcs-server) and exposes the GCS JSON API on port 4443 with an in-memory storage backend.

## Adding this module to your project dependencies

Please run the following command to add the FakeGCSServer module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/fakegcsserver
```

## Usage example

<!--codeinclude-->
[Creating a FakeGCSServer container](../../modules/fakegcsserver/examples_test.go) inside_block:runFakeGCSServerContainer
<!--/codeinclude-->

When connecting with the Google Cloud Storage client library, set the `STORAGE_EMULATOR_HOST` environment variable in your tests:

```go
storageURL, err := ctr.StorageURL(ctx)
// storageURL is e.g. "http://localhost:4443/storage/v1"

t.Setenv("STORAGE_EMULATOR_HOST", storageURL)
```

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The FakeGCSServer module exposes one entrypoint function to create the FakeGCSServer container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "fsouza/fake-gcs-server:1.47")`.

### Container Options

When starting the FakeGCSServer container, you can pass options in a variadic way to configure it.

#### WithScheme

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the HTTP scheme used by the fake-gcs-server. Valid values are `"http"` (default) and `"https"`.

```golang
fakegcsserver.Run(ctx, "fsouza/fake-gcs-server:1.47",
    fakegcsserver.WithScheme("https"),
)
```

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The FakeGCSServer container exposes the following methods:

#### StorageURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the GCS-compatible storage URL for the container. The URL format is `<scheme>://<host>:<port>/storage/v1`, where `scheme` matches the value passed to `WithScheme` (default `"http"`).

```golang
storageURL, err := ctr.StorageURL(ctx)
```

