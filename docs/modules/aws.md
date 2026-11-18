# AWS Emulator

This prototype module starts [AWS Emulator](https://github.com/aws/aws-emulator),
a local implementation of AWS APIs for development and integration tests.

## Running a container

<!--codeinclude-->
[Starting AWS Emulator](../../modules/aws/examples_test.go) inside_block:runAWSContainer
<!--/codeinclude-->

AWS does not currently publish an official container image. The example uses a
pinned build of the **unofficial** [community image](https://github.com/xSAVIKx/aws-emulator-docker),
available for Linux amd64 and arm64. `Run` accepts an image reference so you can
substitute your own image with an `aws-emulator` entrypoint.

The module binds the emulator to `0.0.0.0:9297` inside the container and waits for
its HTTPS health endpoint. `EndpointURL(ctx)` returns the dynamically mapped host
address. Keep the module's container port, `HOME`, and authentication settings
when customizing the request: its file helpers depend on those settings.

## Connecting an AWS SDK client

`Credentials(ctx)` reads the keys generated for this container. `CACertificate(ctx)`
returns its PEM certificate. `TLSConfig(ctx)` trusts that certificate and verifies
its `localhost` identity, including when dialing a remote Docker host. It does not
disable certificate verification. No host AWS credentials or configuration files
are needed.

<!--codeinclude-->
[AWS SDK v2 configuration](../../modules/aws/examples/aws_test.go) inside_block:awsSDKConfig
<!--/codeinclude-->

SDK dependencies and executable STS, S3 and SQS tests live in the separate
`modules/aws/examples` Go module. S3 clients must use path-style addressing.

`WithDisableTelemetry()` disables upstream usage telemetry; the examples use it.

## Prototype limitations

The module exposes the shared AWS API endpoint. Services that return their own
URLs or open extra resource listeners may advertise container-local addresses
or unmapped ports. Those resource endpoints are not rewritten or exposed by
this prototype. The SQS example uses the returned QueueUrl as an identifier
while connecting through the configured SDK endpoint.

AWS API coverage and fidelity vary by service and release. Passing the example
tests does not establish compatibility with every AWS service or behavior.

## Validation

Run `go test -v ./...` in `modules/aws`, then in `modules/aws/examples`.
Both require Docker and access to the pinned community image registry.
