# AWS Emulator

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [AWS Emulator](https://github.com/aws/aws-emulator),
a local implementation of AWS APIs for development and integration tests.
The module starts the emulator with HTTPS and request authentication enabled.

## Adding this module to your project dependencies

Please run the following command to add the AWS Emulator module to your Go dependencies:

```bash
go get github.com/testcontainers/testcontainers-go/modules/aws
```

## Usage example

<!--codeinclude-->
[Creating an AWS Emulator container](../../modules/aws/examples_test.go) inside_block:runAWSContainer
<!--/codeinclude-->

### Configuring the AWS SDK v2

Configure the SDK with the container's endpoint, generated credentials and TLS
configuration. No host AWS credentials or configuration files are needed.

<!--codeinclude-->
[AWS SDK v2 configuration](../../modules/aws/examples/aws_test.go) inside_block:awsSDKConfig
<!--/codeinclude-->

SDK dependencies and executable STS, S3 and SQS tests live in the separate
`modules/aws/examples` Go module. S3 clients must use path-style addressing.

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The AWS Emulator module exposes one entrypoint function to create the AWS Emulator container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
For example: `Run(context.Background(), "ghcr.io/xsavikx/aws-emulator:1.0.509.1-20261118.43")`.

AWS does not currently publish an official AWS Emulator container image. The
example uses a pinned build of the **unofficial**
[community image](https://github.com/xSAVIKx/aws-emulator-docker), available for
Linux amd64 and arm64. You can substitute your own image with an `aws-emulator`
entrypoint that supports the `--bind`, `--port` and `--auth` flags.

The module binds the emulator to `0.0.0.0:9297` inside the container and waits for
its HTTPS health endpoint.

### Container Options

When starting the AWS Emulator container, you can pass options in a variadic way to configure it.

#### WithDisableTelemetry

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`aws.WithDisableTelemetry()` disables AWS Emulator's usage telemetry by setting
`EMULATOR_TELEMETRY=off`. The usage example includes this option.

{% include "../features/common_functional_options_list.md" %}

!!! warning
    Keep the module's container port (`9297`), `HOME` (`/tmp/testcontainers-aws`),
    and authentication settings when customizing the request. Its file helpers
    depend on those settings.

### Container Methods

The AWS Emulator container exposes the following methods:

#### EndpointURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`EndpointURL(ctx)` returns the HTTPS API endpoint, including the dynamically
mapped host port for container port `9297`.

#### Credentials

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`Credentials(ctx)` returns a `Credentials` value containing the `AccessKeyID`
and `SecretAccessKey` generated for this emulator instance. Use these keys to
configure the SDK's credentials provider. They cannot authenticate to AWS.

#### CACertificate

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`CACertificate(ctx)` returns this instance's generated certificate in PEM format.
Use it to configure certificate trust for clients outside Go.

#### TLSConfig

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`TLSConfig(ctx)` returns a `*tls.Config` that trusts this instance's certificate
and verifies its `localhost` identity, including when dialing a remote Docker
host. It does not disable certificate verification. The SDK example above uses
this configuration with an HTTP client.

### Limitations

The module exposes the shared AWS API endpoint. Services that return their own
URLs or open extra resource listeners may advertise container-local addresses
or unmapped ports. Those resource endpoints are not rewritten or exposed by
this prototype. The SQS example uses the returned QueueUrl as an identifier
while connecting through the configured SDK endpoint.

AWS API coverage and fidelity vary by service and release. Passing the example
tests does not establish compatibility with every AWS service or behavior.
