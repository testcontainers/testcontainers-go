# LocalStack

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.18.0"><span class="tc-version">:material-tag: v0.18.0</span></a>

## Introduction

The Testcontainers module for [LocalStack](http://localstack.cloud/) is _"a fully functional local AWS cloud stack"_, to develop and test your cloud and serverless apps without actually using the cloud.

## Adding this module to your project dependencies

Please run the following command to add the LocalStack module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/localstack
```

## Usage example

Running LocalStack as a stand-in for multiple AWS services during a test:

<!--codeinclude-->
[Creating a LocalStack container](../../modules/localstack/examples_test.go) inside_block:runLocalstackContainer
<!--/codeinclude-->

Environment variables listed in [Localstack's README](https://github.com/localstack/localstack#configurations) may be used to customize Localstack's configuration. 
Use the `testcontainers.WithEnv` option when creating the `LocalStackContainer` to apply those variables.

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The LocalStack module exposes one single function to create the LocalStack container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*LocalstackContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Localstack container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "localstack:1.4.0")`.

{% include "../features/common_functional_options.md" %}

## Accessing hostname-sensitive services

Some Localstack APIs, such as SQS, require the container to be aware of the hostname that it is accessible on - for example, for construction of queue URLs in responses.

Testcontainers will inform Localstack of the best hostname automatically, using an environment variable for that:

* for Localstack versions 0.10.0 and above, the `HOSTNAME_EXTERNAL` environment variable will be set to hostname in the container request.
* for Localstack versions 2.0.0 and above, the `LOCALSTACK_HOST` environment variable will be set to the hostname in the container request.

Once the variable is set:

* when running the Localstack container directly without a custom network defined, it is expected that all calls to the container will be from the test host. As such, the container address will be used (typically localhost or the address where the Docker daemon is running).

* when running the Localstack container [with a custom network defined](/features/networking/#advanced-networking), it is expected that all calls to the container will be **from other containers on that network**. `HOSTNAME_EXTERNAL` will be set to the *last* network alias that has been configured for the Localstack container.

    <!--codeinclude-->
    [Localstack container running with a custom network](../../modules/localstack/examples_test.go) inside_block:localstackWithNetwork
    <!--/codeinclude-->

* Other usage scenarios, such as where the Localstack container is used from both the test host and containers on a custom network, are not automatically supported. If you have this use case, you should set `HOSTNAME_EXTERNAL` manually.

## Obtaining a client using the AWS SDK for Go

You can use the AWS SDK for Go to create a client for the LocalStack container. The following examples show how to create a client for the S3 service, using both the SDK v1 and v2.

### Using the AWS SDK v1

<!--codeinclude-->
[AWS SDK v1](../../modules/localstack/v1/s3_test.go) inside_block:awsSDKClientV1
<!--/codeinclude-->

For further reference on the SDK v1, please check out the AWS docs [here](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html).

### Using the AWS SDK v2

<!--codeinclude-->
[EndpointResolver](../../modules/localstack/v2/s3_test.go) inside_block:awsResolverV2
[AWS SDK v2](../../modules/localstack/v2/s3_test.go) inside_block:awsSDKClientV2
<!--/codeinclude-->

For further reference on the SDK v2, please check out the AWS docs [here](https://aws.github.io/aws-sdk-go-v2/docs/getting-started)

## Testing the module

The module includes unit and integration tests that can be run from its source code. To run the tests please execute the following command:

```
cd modules/localstack
make test
```

!!!info
	At this moment, the tests for the module include tests for the S3 service, only. They live in two different Go packages of the LocalStack module,
    `v1` and `v2`, where it'll be easier to add more examples for the rest of services.
