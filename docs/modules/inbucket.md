# Inbucket

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

## Introduction

The Testcontainers module for Inbucket.

## Adding this module to your project dependencies

Please run the following command to add the Inbucket module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/inbucket
```

## Usage example

<!--codeinclude-->
[Creating an Inbucket container](../../modules/inbucket/examples_test.go) inside_block:runInbucketContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Inbucket module exposes one entrypoint function to create the Inbucket container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*InbucketContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Inbucket container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "inbucket/inbucket:sha-2d409bb")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Inbucket container exposes the following methods:

#### SmtpConnection

This method returns the connection string to connect to the Inbucket container SMTP service, using the `2500` port.

<!--codeinclude-->
[Get smtp connection string](../../modules/inbucket/inbucket_test.go) inside_block:smtpConnection
<!--/codeinclude-->

#### WebInterface

This method returns the connection string to connect to the Inbucket container web interface, using the `9000` port.

<!--codeinclude-->
[Get web interface connection string](../../modules/inbucket/inbucket_test.go) inside_block:webInterface
<!--/codeinclude-->
