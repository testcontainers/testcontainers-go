# Inbucket

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Inbucket.

## Adding this module to your project dependencies

Please run the following command to add the Inbucket module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/inbucket
```

## Usage example

<!--codeinclude-->
[Creating a Inbucket container](../../modules/inbucket/examples_test.go) inside_block:runInbucketContainer
<!--/codeinclude-->

## Module reference

The Inbucket module exposes one entrypoint function to create the Inbucket container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*InbucketContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Inbucket container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Inbucket Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Inbucket. E.g. `testcontainers.WithImage("inbucket/inbucket:sha-2d409bb")`.

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
