# S3Mock

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [S3Mock](https://github.com/adobe/S3Mock) — a popular open-source library by Adobe that mocks the AWS S3 API for use in tests. S3Mock runs as a lightweight Docker container and fully implements the S3 API, allowing tests to interact with S3 buckets and objects without real AWS credentials or network access.

## Adding this module to your project dependencies

Please run the following command to add the S3Mock module to your Go dependencies:

```bash
go get github.com/testcontainers/testcontainers-go/modules/s3mock
```

## Usage example

<!--codeinclude-->
[Creating a S3Mock container](../../modules/s3mock/examples_test.go) inside_block:runS3MockContainer
<!--/codeinclude-->

### Configuring the AWS SDK v2

S3Mock fully implements the AWS S3 API. Point the AWS SDK v2 at the container's `EndpointURL` to use it in tests — no real credentials are required.

First implement a custom `EndpointResolverV2` that routes all S3 calls to the container:

<!--codeinclude-->
[Endpoint resolver](../../modules/s3mock/s3mock_test.go) inside_block:endpointResolver
<!--/codeinclude-->

Then create the S3 client:

<!--codeinclude-->
[AWS SDK v2 client setup](../../modules/s3mock/examples_test.go) inside_block:awsClientSetup
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The S3Mock module exposes one entrypoint function to create the S3Mock container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "adobe/s3mock:3.9")`.

### Container Options

When starting the S3Mock container, you can pass options in a variadic way to configure it.

#### WithInitialBuckets

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use `WithInitialBuckets` to pre-create one or more S3 buckets when the container starts:

<!--codeinclude-->
[WithInitialBuckets](../../modules/s3mock/s3mock_test.go) inside_block:withInitialBuckets
<!--/codeinclude-->

{% include "../features/common_functional_options_list.md" %}

### Container Methods

#### EndpointURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the HTTP endpoint URL for the S3Mock container (mapped from container port 9090). Use this URL as the base endpoint when configuring the AWS SDK.

<!--codeinclude-->
[Get HTTP endpoint URL](../../modules/s3mock/s3mock_test.go) inside_block:endpointURL
<!--/codeinclude-->

#### HTTPSEndpointURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the HTTPS endpoint URL for the S3Mock container (mapped from container port 9191).

<!--codeinclude-->
[Get HTTPS endpoint URL](../../modules/s3mock/s3mock_test.go) inside_block:httpsEndpointURL
<!--/codeinclude-->
