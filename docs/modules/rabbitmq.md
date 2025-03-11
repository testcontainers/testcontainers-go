# RabbitMQ

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.25.0"><span class="tc-version">:material-tag: v0.25.0</span></a>

## Introduction

The Testcontainers module for RabbitMQ.

## Adding this module to your project dependencies

Please run the following command to add the RabbitMQ module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/rabbitmq
```

## Usage example

<!--codeinclude-->
[Creating a RabbitMQ container](../../modules/rabbitmq/examples_test.go) inside_block:runRabbitMQContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The RabbitMQ module exposes one entrypoint function to create the RabbitMQ container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*RabbitMQContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the RabbitMQ container, you can pass options in a variadic way to configure it. All these options will be automatically rendered into the RabbitMQ's custom configuration file, located at `/etc/rabbitmq/rabbitmq-custom.conf`.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "rabbitmq:3.7.25-management-alpine")`.

!!!warning
    From [https://hub.docker.com/_/rabbitmq](https://hub.docker.com/_/rabbitmq): "As of RabbitMQ 3.9, all of the docker-specific variables listed below are deprecated and no longer used. Please use a configuration file instead; visit [rabbitmq.com/configure](https://rabbitmq.com/configure) to learn more about the configuration file. For a starting point, the 3.8 images will print out the config file it generated from supplied environment variables."

    - RABBITMQ_DEFAULT_PASS_FILE
    - RABBITMQ_DEFAULT_USER_FILE
    - RABBITMQ_MANAGEMENT_SSL_CACERTFILE
    - RABBITMQ_MANAGEMENT_SSL_CERTFILE
    - RABBITMQ_MANAGEMENT_SSL_DEPTH
    - RABBITMQ_MANAGEMENT_SSL_FAIL_IF_NO_PEER_CERT
    - RABBITMQ_MANAGEMENT_SSL_KEYFILE
    - RABBITMQ_MANAGEMENT_SSL_VERIFY
    - RABBITMQ_SSL_CACERTFILE
    - RABBITMQ_SSL_CERTFILE
    - RABBITMQ_SSL_DEPTH
    - RABBITMQ_SSL_FAIL_IF_NO_PEER_CERT
    - RABBITMQ_SSL_KEYFILE
    - RABBITMQ_SSL_VERIFY
    - RABBITMQ_VM_MEMORY_HIGH_WATERMARK

{% include "../features/common_functional_options.md" %}

#### Startup Commands for RabbitMQ

The RabbitMQ module includes several test implementations of the `testcontainers.Executable` interface: Binding, Exchange, OperatorPolicy, Parameter, Permission, Plugin, Policy, Queue, User, VirtualHost and VirtualHostLimit. You could use them as reference to understand how the startup commands are generated, but please consider this test implementation could not be complete for your use case.

You could use this feature to run a custom script, or to run a command that is not supported by the module. RabbitMQ examples of this could be:

- Enable plugins
- Add virtual hosts and virtual hosts limits
- Add exchanges
- Add queues
- Add bindings
- Add policies
- Add operator policies
- Add parameters
- Add permissions
- Add users

Please refer to the RabbitMQ documentation to build your own commands.

<!--codeinclude-->
[Add Virtual Hosts](../../modules/rabbitmq/rabbitmq_test.go) inside_block:addVirtualHosts
[Add Exchanges](../../modules/rabbitmq/rabbitmq_test.go) inside_block:addExchanges
[Add Queues](../../modules/rabbitmq/rabbitmq_test.go) inside_block:addQueues
[Add Bindings](../../modules/rabbitmq/rabbitmq_test.go) inside_block:addBindings
[Add Policies](../../modules/rabbitmq/rabbitmq_test.go) inside_block:addPolicies
[Add Permissions](../../modules/rabbitmq/rabbitmq_test.go) inside_block:addPermissions
[Add Users](../../modules/rabbitmq/rabbitmq_test.go) inside_block:addUsers
[Enabling Plugins](../../modules/rabbitmq/rabbitmq_test.go) inside_block:enablePlugins
<!--/codeinclude-->

#### Default Admin

If you need to set the username and/or password for the admin user, you can use the `WithAdminUsername(username string)` and `WithAdminPassword(pwd string)` options.

!!!info
    By default, the admin username is `guest` and the password is `guest`.

#### SSL settings

In the case you need to enable SSL, you can use the `WithSSL(settings SSLSettings)` option. This option will enable SSL with the passed settings:

<!--codeinclude-->
[Enabling SSL](../../modules/rabbitmq/examples_test.go) inside_block:enableSSL
<!--/codeinclude-->

You'll find a log entry similar to this one in the container logs:

```
2023-09-13 13:05:10.213 [info] <0.548.0> started TLS (SSL) listener on [::]:5671
```

### Container Methods

The RabbitMQ container exposes the following methods:

#### AMQP URLs

The RabbitMQ container exposes two methods to retrieve the AMQP URLs in order to connect to the RabbitMQ instance using AMQP clients:

- `AmqpURL()`, returns the AMQP URL.
- `AmqpsURL()`, returns the AMQPS URL.

#### HTTP management URLs

The RabbitMQ container exposes two methods to retrieve the HTTP URLs for management:

- `HttpURL()`, returns the management URL over HTTP.
- `HttpsURL()`, returns the management URL over HTTPS.
