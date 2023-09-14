# RabbitMQ

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

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

## Module reference

The RabbitMQ module exposes one entrypoint function to create the RabbitMQ container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RabbitMQContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the RabbitMQ container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different RabbitMQ Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for RabbitMQ. E.g. `testcontainers.WithImage("rabbitmq:3.7.25-management-alpine")`.

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

#### Wait Strategies

If you need to set a different wait strategy for RabbitMQ, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for RabbitMQ.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for RabbitMQ, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

#### Startup Commands

!!!info
    Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Testcontainers exposes the `WithStartupCommand(e ...Executable)` option to run arbitrary commands in the container right after it's started.

!!!info
    To better understand how this feature works, please read the [Create containers: Lifecycle Hooks](../../features/creating_container/#lifecycle-hooks) documentation.

It also exports an `Executable` interface, defining one single method: `AsCommand()`, which returns a slice of strings to represent the command and positional arguments to be executed in the container.

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the RabbitMQ container is started.

!!!info
    The RabbitMQ module includes a couple of test implementations of the `Executable` interface: Binding, Exchange, OperatorPolicy, Parameter, Permission, Plugin, Policy, Queue, User, VirtualHost and VirtualHostLimit. You could use them as reference, but consider the implementation could not be complete for your use case.

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

#### Custom Configuration

If you need to set a custom configuration, you can use `WithConfig(filePath string)` option to pass the path to a custom configuration file in Sysctl format, or user `WithConfigErlang(filePath string)` to pass the path to a custom configuration file in Erlang format.

!!!warning
    The `WithConfig` option doesn't work with RabbitMQ &lt; 3.7, and it's recommended for RabbitMQ &gt;= 3.7.

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
