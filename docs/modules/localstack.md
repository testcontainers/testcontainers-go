# LocalStack

The Testcontainers module for [LocalStack](http://localstack.cloud/) is _"a fully functional local AWS cloud stack"_, to develop and test your cloud and serverless apps without actually using the cloud.

## Adding this module to your project dependencies

Please run the following command to add the LocalStack module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/localstack
```

## Usage example

Running LocalStack as a stand-in for multiple AWS services during a test:

<!--codeinclude-->
[Creating a LocalStack container](../../modules/localstack/v1/s3_test.go) inside_block:localStackCreateContainer
<!--/codeinclude-->

Environment variables listed in [Localstack's README](https://github.com/localstack/localstack#configurations) may be used to customize Localstack's configuration. 
Use the `OverrideContainerRequest` option when creating the `LocalStackContainer` to apply configuration settings.

## Creating a client using the AWS SDK for Go

### Version 1

<!--codeinclude-->
[Test for a LocalStack container, usinv AWS SDK v1](../../modules/localstack/v1/s3_test.go) inside_block:awsSDKClientV1
<!--/codeinclude-->

For further reference on the SDK v1, please check out the AWS docs [here](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html).

### Version 2

<!--codeinclude-->
[Test for a LocalStack container, usinv AWS SDK v2](../../modules/localstack/v2/s3_test.go) inside_block:awsSDKClientV2
<!--/codeinclude-->

For further reference on the SDK v2, please check out the AWS docs [here](https://aws.github.io/aws-sdk-go-v2/docs/getting-started)

## `HOSTNAME_EXTERNAL` and hostname-sensitive services

Some Localstack APIs, such as SQS, require the container to be aware of the hostname that it is accessible on - for example, for construction of queue URLs in responses.

Testcontainers will inform Localstack of the best hostname automatically, using the `HOSTNAME_EXTERNAL` environment variable:

* when running the Localstack container directly without a custom network defined, it is expected that all calls to the container will be from the test host. As such, the container address will be used (typically localhost or the address where the Docker daemon is running).

    <!--codeinclude-->
    [Localstack container running without a custom network](../../modules/localstack/localstack_test.go) inside_block:withoutNetwork
    <!--/codeinclude-->

* when running the Localstack container [with a custom network defined](/features/networking/#advanced-networking), it is expected that all calls to the container will be **from other containers on that network**. `HOSTNAME_EXTERNAL` will be set to the *last* network alias that has been configured for the Localstack container.

    <!--codeinclude-->
    [Localstack container running with a custom network](../../modules/localstack/localstack_test.go) inside_block:withNetwork
    <!--/codeinclude-->

* Other usage scenarios, such as where the Localstack container is used from both the test host and containers on a custom network are not automatically supported. If you have this use case, you should set `HOSTNAME_EXTERNAL` manually.

## Module reference

The LocalStack module exposes one single function to create the LocalStack container, and this function receives two parameters:

```golang
func StartContainer(ctx context.Context, overrideReq OverrideContainerRequestOption) (*LocalStackContainer, error)
```

- `context.Context`
- `OverrideContainerRequestOption`

### OverrideContainerRequestOption

The `OverrideContainerRequestOption` functional option represents a way to override the default LocalStack container request:

<!--codeinclude-->
[Default container request](../../modules/localstack/localstack.go) inside_block:defaultContainerRequest
<!--/codeinclude-->

With simply passing your own instance of an `OverrideContainerRequestOption` type to the `StartContainer` function, you'll be able to configure the LocalStack container with your own needs, as this new container request will be merged with the original one.

In the following example you check how it's possible to set certain environment variables that are needed by the tests, the most important of them the AWS services you want to use. Besides, the container runs in a separate Docker network with an alias:

<!--codeinclude-->
[Overriding the default container request](../../modules/localstack/localstack_test.go) inside_block:withNetwork
<!--/codeinclude-->

If you do not need to override the container request, you can pass `nil` or the `NoopOverrideContainerRequest` instance, which is exposed as a helper for this reason.

<!--codeinclude-->
[Skip overriding the default container request](../../modules/localstack/localstack_test.go) inside_block:noopOverrideContainerRequest
<!--/codeinclude-->

## Testing the module

The module includes unit and integration tests that can be run from its source code. To run the tests please execute the following command:

```
cd modules/localstack
make test
```

!!!info

	At this moment, the tests for the module include tests for the S3 service, only. They live in two different Go packages of the LocalStack module,
    `v1` and `v2`, where it'll be easier to add more examples for the rest of services.
