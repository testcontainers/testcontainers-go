# PapercutSMTP

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [Papercut SMTP](https://github.com/ChangemakerStudios/Papercut-SMTP), a lightweight SMTP server with a built-in web UI for capturing and inspecting outbound emails during testing. No messages are relayed to real recipients — all sent mail is held in memory and viewable through the web interface.

## Adding this module to your project dependencies

Please run the following command to add the PapercutSMTP module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/papercutsmtp
```

## Usage example

<!--codeinclude-->
[Creating a PapercutSMTP container](../../modules/papercutsmtp/examples_test.go) inside_block:runPapercutSMTPContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The PapercutSMTP module exposes one entrypoint function to create the PapercutSMTP container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "changemakerstudiosus/papercut-smtp:latest")`.

### Container Options

When starting the PapercutSMTP container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The PapercutSMTP container exposes the following methods:

#### SMTPEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the `host:port` endpoint to connect to the SMTP server on port `25`.

<!--codeinclude-->
[Get SMTP endpoint](../../modules/papercutsmtp/papercutsmtp_test.go) inside_block:smtpEndpoint
<!--/codeinclude-->

#### HTTPURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the URL for the Papercut SMTP web UI on port `37408`, where captured emails can be inspected.

<!--codeinclude-->
[Get HTTP URL](../../modules/papercutsmtp/papercutsmtp_test.go) inside_block:httpURL
<!--/codeinclude-->
