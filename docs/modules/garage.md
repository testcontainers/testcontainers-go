# Garage

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [Garage](https://garagehq.deuxfleurs.fr/), an S3-compatible object store.
The module starts a single-node server and creates an access key and bucket before accepting S3 requests.
Use an S3 client with path-style addressing, as shown in the example below.

## Adding this module to your project dependencies

Please run the following command to add the Garage module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/garage
```

## Usage example

<!--codeinclude-->
[Creating a Garage container](../../modules/garage/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Garage module exposes one entrypoint function to create the Garage container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
For example: `Run(context.Background(), "dxflrs/garage:v2.4.1")`.

Garage v2.3.0 or later is required for the `--single-node` and `--default-bucket` bootstrap flags.

### Container Options

When starting the Garage container, you can pass options in a variadic way to configure it.

#### WithCredentials

`garage.WithCredentials(accessKey, secretKey)` sets the initial S3 credentials.
The defaults are `testcontainers` and `testcontainers-secret`, respectively.
The access key must contain at least 8 bytes and the secret key at least 16 bytes.
These credentials are intended for tests. The initial key can create additional buckets using the S3 API.

#### WithBucket

`garage.WithBucket("my-bucket")` changes the initial bucket name from `testcontainers`.
The bucket must have a valid S3 bucket name.

#### WithRegion

`garage.WithRegion("eu-west-1")` changes the S3 region from `garage`.
Configure your S3 client to use the same region.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Garage container exposes the following methods:

#### ConnectionString

`ConnectionString(ctx)` returns the S3 HTTP endpoint, including the mapped port.

#### AccessKey and SecretKey

`AccessKey()` and `SecretKey()` return the initial S3 credentials.

#### Bucket

`Bucket()` returns the name of the bucket created at startup.

#### Region

`Region()` returns the configured S3 region.

### Stopping the container

Garage can take the full stop timeout to exit after serving S3 requests. If graceful shutdown is unnecessary
for your tests, use `testcontainers.TerminateContainer(ctr, testcontainers.StopTimeout(0))`
or `testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))`.
