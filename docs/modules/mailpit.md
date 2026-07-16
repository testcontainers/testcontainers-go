# Mailpit

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Mailpit.

[Mailpit](https://mailpit.axllent.org/) is a fast, multi-platform email testing tool. It provides an SMTP server that captures outgoing emails and exposes them through a web UI and REST API, making it ideal for asserting email sending behaviour in integration tests.

## Adding this module to your project dependencies

Please run the following command to add the Mailpit module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mailpit
```

## Usage example

<!--codeinclude-->
[Creating a Mailpit container](../../modules/mailpit/examples_test.go) inside_block:runMailpitContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Mailpit module exposes one entrypoint function to create the Mailpit container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "axllent/mailpit:v1.20")`.

### Container Options

When starting the Mailpit container, you can pass options in a variadic way to configure it.

#### WithSMTPAuth

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`WithSMTPAuth` sets the SMTP authentication credentials for the Mailpit container, using the `MP_SMTP_AUTH_USERNAME` and `MP_SMTP_AUTH_PASSWORD` environment variables.

<!--codeinclude-->
[WithSMTPAuth](../../modules/mailpit/mailpit_test.go) inside_block:withSMTPAuth
<!--/codeinclude-->

#### WithMessageLimit

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`WithMessageLimit` sets the maximum number of messages to store in Mailpit, using the `MP_MAX_MESSAGES` environment variable. When the limit is reached the oldest messages are automatically deleted.

<!--codeinclude-->
[WithMessageLimit](../../modules/mailpit/mailpit_test.go) inside_block:withMessageLimit
<!--/codeinclude-->

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Mailpit container exposes the following methods:

#### SMTPEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`SMTPEndpoint` returns the `host:port` connection string for the Mailpit SMTP server, using the `1025` port.

<!--codeinclude-->
[Get SMTP endpoint](../../modules/mailpit/mailpit_test.go) inside_block:smtpEndpoint
<!--/codeinclude-->

#### HTTPURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`HTTPURL` returns the URL for the Mailpit web interface and REST API, using the `8025` port. The REST API is available at `/api/v1/messages` and other endpoints under `/api/v1/`.

<!--codeinclude-->
[Get HTTP URL](../../modules/mailpit/mailpit_test.go) inside_block:httpURL
<!--/codeinclude-->

