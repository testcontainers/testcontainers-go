# Apache ActiveMQ Artemis

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.23.0"><span class="tc-version">:material-tag: v0.23.0</span></a>

## Introduction

The Testcontainers module for Artemis.

## Adding this module to your project dependencies

Please run the following command to add the Artemis module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/artemis
```

## Usage example

<!--codeinclude-->
[Creating to an Artemis container](../../modules/artemis/examples_test.go) inside_block:runArtemisContainer
<!--/codeinclude-->

<!--codeinclude-->
[Connecting to an Artemis container](../../modules/artemis/examples_test.go) inside_block:connectToArtemisContainer
<!--/codeinclude-->

## Module reference

The Artemis module exposes one entrypoint function to create the Artemis container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Artemis container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Artemis Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Artemis. E.g. `testcontainers.WithImage("docker.io/apache/activemq-artemis:2.30.0")`.

#### Wait Strategies

If you need to set a different wait strategy for Artemis, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for Artemis.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Credentials

If you need to change the default admin credentials (i.e. `artemis:artemis`) use `WithCredentials`.

<!--codeinclude-->
[With credentials](../../modules/artemis/artemis_test.go) inside_block:withCredentials
<!--/codeinclude-->

#### Enable Anonymous login

If you need to enable anonymous logins (which are disabled by default) use `WithAnonymousLogin`.

<!--codeinclude-->
[With Anonymous Login](../../modules/artemis/artemis_test.go) inside_block:withAnonymousLogin
<!--/codeinclude-->

#### Custom Arguments

If you need to pass custom arguments to the `artemis create` command, use `WithExtraArgs`.
The default is `--http-host 0.0.0.0 --relax-jolokia`.
Setting this value will override the default.

!!!info
    Please see the documentation on `artemis create` for the available options here: [https://activemq.apache.org/components/artemis/documentation/latest/using-server.html#options](https://activemq.apache.org/components/artemis/documentation/latest/using-server.html#options)

<!--codeinclude-->
[With Extra Arguments](../../modules/artemis/artemis_test.go) inside_block:withExtraArgs
<!--/codeinclude-->

#### Docker type modifiers

If you need an advanced configuration for Artemis, you can leverage the following Docker type modifiers:

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

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the Artemis container is started.

### Container Methods

The Artemis container exposes the following methods:

#### User

User returns the administrator username.

<!--codeinclude-->
[Retrieving the Administrator User](../../modules/artemis/examples_test.go) inside_block:containerUser
<!--/codeinclude-->

#### Password

Password returns the administrator password.

<!--codeinclude-->
[Retrieving the Administrator Password](../../modules/artemis/examples_test.go) inside_block:containerPassword
<!--/codeinclude-->

#### BrokerEndpoint

BrokerEndpoint returns the host:port for the combined protocols endpoint.

<!--codeinclude-->
[Get broker endpoint](../../modules/artemis/artemis_test.go) inside_block:brokerEndpoint
<!--/codeinclude-->

#### ConsoleURL

ConsoleURL returns the URL for the management console.

<!--codeinclude-->
[Get console URL](../../modules/artemis/artemis_test.go) inside_block:consoleURL
<!--/codeinclude-->
