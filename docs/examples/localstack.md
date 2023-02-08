# LocalStack

The Testcontainers module for [LocalStack](http://localstack.cloud/) is _"a fully functional local AWS cloud stack"_, to develop and test your cloud and serverless apps without actually using the cloud.

## Adding this module to your project dependencies

Please run the following command to add the LocalStack module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/examples/localstack
```

## Usage example

Running LocalStack as a stand-in for multiple AWS services during a test:

<!--codeinclude-->
[Creating a LocalStack container](../../examples/localstack/v1/s3_test.go) inside_block:localStackCreateContainer
<!--/codeinclude-->

Environment variables listed in [Localstack's README](https://github.com/localstack/localstack#configurations) may be used to customize Localstack's configuration. 
Use the `OverrideContainerRequest` option when creating the `LocalStackContainer` to apply configuration settings.

## Creating a client using the AWS SDK for Go

### Version 1

<!--codeinclude-->
[Test for a LocalStack container, usinv AWS SDK v1](../../examples/localstack/v1/s3_test.go) inside_block:awsSDKClientV1
<!--/codeinclude-->

For further reference on the SDK v1, please check out the AWS docs [here](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html).

### Version 2

<!--codeinclude-->
[Test for a LocalStack container, usinv AWS SDK v2](../../examples/localstack/v2/s3_test.go) inside_block:awsSDKClientV2
<!--/codeinclude-->

For further reference on the SDK v2, please check out the AWS docs [here](https://aws.github.io/aws-sdk-go-v2/docs/getting-started)

## `HOSTNAME_EXTERNAL` and hostname-sensitive services

Some Localstack APIs, such as SQS, require the container to be aware of the hostname that it is accessible on - for example, for construction of queue URLs in responses.

Testcontainers will inform Localstack of the best hostname automatically, using the `HOSTNAME_EXTERNAL` environment variable:

* when running the Localstack container directly without a custom network defined, it is expected that all calls to the container will be from the test host. As such, the container address will be used (typically localhost or the address where the Docker daemon is running).

    <!--codeinclude-->
    [Localstack container running without a custom network](../../examples/localstack/localstack_legacy_mode_test.go) inside_block:withoutNetwork
    <!--/codeinclude-->

* when running the Localstack container [with a custom network defined](/features/networking/#advanced-networking), it is expected that all calls to the container will be **from other containers on that network**. `HOSTNAME_EXTERNAL` will be set to the *last* network alias that has been configured for the Localstack container.

    <!--codeinclude-->
    [Localstack container running with a custom network](../../examples/localstack/localstack_test.go) inside_block:withNetwork
    <!--/codeinclude-->

* Other usage scenarios, such as where the Localstack container is used from both the test host and containers on a custom network are not automatically supported. If you have this use case, you should set `HOSTNAME_EXTERNAL` manually.

## Module reference

The LocalStack module exposes one single function to create containers, and this function receives three parameters:

- `context.Context`
- `OverrideContainerRequestOption`
- a variadic argument of `LocalStackContainerOption`

### OverrideContainerRequestOption

The `OverrideContainerRequestOption` functional option represents a way to override the default LocalStack container request:

<!--codeinclude-->
[Default container request](../../examples/localstack/localstack.go) inside_block:defaultContainerRequest
<!--/codeinclude-->

With simply passing your own instance of an `OverrideContainerRequestOption` type to the `StartContainer` function, you'll be able to configure the LocalStack container with your own needs.

In the following example you check how it's possible to set certain environment variables that are needed by the tests. Besides, the container runs in a separate Docker network with an alias:

<!--codeinclude-->
[Overriding the default container request](../../examples/localstack/localstack_test.go) inside_block:withNetwork
<!--/codeinclude-->

If you do not need to override the container request, you can pass `nil` or the `NoopOverrideContainerRequest` instance, which is exposed as a helper for this reason.

<!--codeinclude-->
[Skip overriding the default container request](../../examples/localstack/localstack_test.go) inside_block:noopOverrideContainerRequest
<!--/codeinclude-->

### LocalStackContainerOption, variadic argument

#### WithLegacyMode

The `WithLegacyMode` functional option represents a way to force LocalStack to run in legacy mode.

<!--codeinclude-->
[Forcing legacy mode](../../examples/localstack/localstack_legacy_mode_test.go) inside_block:forceLegacyMode
<!--/codeinclude-->

#### WithServices

The `WithServices` functional option represents a way to pass as many AWS services as needed, in a variadic manner:

<!--codeinclude-->
[Passing AWS services to LocalStack](../../examples/localstack/v1/s3_test.go) inside_block:localStackCreateContainer
<!--/codeinclude-->

Using this method will internally populate the `SERVICES` environment variable, alongside the exposed ports of each service.

### Available AWS Services

The LocalStack module supports the following AWS services:

- APIGateway
- CloudFormation
- CloudWatch
- CloudWatchLogs
- DynamoDB
- DynamoDBStreams
- EC2
- Firehose
- IAM
- KMS
- Kinesis
- Lambda
- Redshift
- Route53
- S3
- SES
- SNS
- SQS
- SSM
- STS
- SecretsManager
- StepFunctions

!!!info

	At this moment, the tests for the module include tests for the S3 service, only. They live in two different Go packages of the LocalStack module,
    `v1` and `v2`, where it'll be easier to add more examples for the rest of services.
