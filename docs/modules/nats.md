# NATS

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

## Introduction

The Testcontainers module for NATS.

## Adding this module to your project dependencies

Please run the following command to add the NATS module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/nats
```

## Usage example

<!--codeinclude-->
[Creating a NATS container](../../modules/nats/examples_test.go) inside_block:runNATSContainer
<!--/codeinclude-->

## Module reference

The NATS module exposes one entrypoint function to create the NATS container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*NATSContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the NATS container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different NATS Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for NATS. E.g. `testcontainers.WithImage("nats:2.9")`.

#### Wait Strategies

If you need to set a different wait strategy for NATS, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for NATS.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for NATS, you can leverage the following Docker type modifiers:

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

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the NATS container is started.

#### Set username and password

If you need to set different credentials, you can use `WithUsername` and `WithPassword`
options. By default, the username, the password are not set. To establish the connection with the NATS container:

<!--codeinclude-->
[Connect using the credentials](../../modules/nats/examples_test.go) inside_block:natsConnect
<!--/codeinclude-->

### Container Methods

The NATS container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the NATS container, using the default `4222` port.
It's possible to pass extra parameters to the connection string, in a variadic way.

<!--codeinclude-->
[Get connection string](../../modules/nats/nats_test.go) inside_block:connectionString
<!--/codeinclude-->
