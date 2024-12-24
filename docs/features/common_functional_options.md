#### Image Substitutions

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.26.0"><span class="tc-version">:material-tag: v0.26.0</span></a>

In more locked down / secured environments, it can be problematic to pull images from Docker Hub and run them without additional precautions.

An image name substitutor converts a Docker image name, as may be specified in code, to an alternative name. This is intended to provide a way to override image names, for example to enforce pulling of images from a private registry.

_Testcontainers for Go_ exposes an interface to perform this operations: `ImageSubstitutor`, and a No-operation implementation to be used as reference for custom implementations:

<!--codeinclude-->
[Image Substitutor Interface](../../options.go) inside_block:imageSubstitutor
[Noop Image Substitutor](../../container_test.go) inside_block:noopImageSubstitutor
<!--/codeinclude-->

Using the `WithImageSubstitutors` options, you could define your own substitutions to the container images. E.g. adding a prefix to the images so that they can be pulled from a Docker registry other than Docker Hub. This is the usual mechanism for using Docker image proxies, caches, etc.

#### WithEnv

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

If you need to either pass additional environment variables to a container or override them, you can use `testcontainers.WithEnv` for example:

```golang
postgres, err = postgresModule.Run(ctx, "postgres:15-alpine", testcontainers.WithEnv(map[string]string{"POSTGRES_INITDB_ARGS": "--no-sync"}))
```

#### WithHostPortAccess

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.31.0"><span class="tc-version">:material-tag: v0.31.0</span></a>

If you need to access a port that is already running in the host, you can use `testcontainers.WithHostPortAccess` for example:

```golang
postgres, err = postgresModule.Run(ctx, "postgres:15-alpine", testcontainers.WithHostPortAccess(8080))
```

To understand more about this feature, please read the [Exposing host ports to the container](/features/networking/#exposing-host-ports-to-the-container) documentation.

#### WithLogConsumers

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

If you need to consume the logs of the container, you can use `testcontainers.WithLogConsumers` with a valid log consumer. An example of a log consumer is the following:

```golang
type TestLogConsumer struct {
	Msgs []string
}

func (g *TestLogConsumer) Accept(l Log) {
	g.Msgs = append(g.Msgs, string(l.Content))
}
```

#### WithLogger

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

If you need to either pass logger to a container, you can use `testcontainers.WithLogger`.

!!!info
	Consider calling this before other "With" functions as these may generate logs.

In this example we also use `TestLogger` which writes to the passed in `testing.TB` using `Logf`.
The result is that we capture all logging from the container into the test context meaning its
hidden behind `go test -v` and is associated with the relevant test, providing the user with
useful context instead of appearing out of band.

```golang
func TestHandler(t *testing.T) {
    logger := TestLogger(t)
    ctr, err := postgresModule.Run(ctx, "postgres:15-alpine", testcontainers.WithLogger(logger))
    CleanupContainer(t, ctr)
    require.NoError(t, err)
    // Do something with container.
}
```

Please read the [Following Container Logs](/features/follow_logs) documentation for more information about creating log consumers.

#### Wait Strategies

If you need to set a different wait strategy for the container, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Startup Commands

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.25.0"><span class="tc-version">:material-tag: v0.25.0</span></a>

Testcontainers exposes the `WithStartupCommand(e ...Executable)` option to run arbitrary commands in the container right after it's started.

!!!info
    To better understand how this feature works, please read the [Create containers: Lifecycle Hooks](/features/creating_container/#lifecycle-hooks) documentation.

It also exports an `Executable` interface, defining the following methods:

- `AsCommand()`, which returns a slice of strings to represent the command and positional arguments to be executed in the container;
- `Options()`, which returns the slice of functional options with the Docker's ExecConfigs used to create the command in the container (the working directory, environment variables, user executing the command, etc) and the possible output format (Multiplexed).

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the container is started.

#### Ready Commands

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

Testcontainers exposes the `WithAfterReadyCommand(e ...Executable)` option to run arbitrary commands in the container right after it's ready, which happens when the defined wait strategies have finished with success.

!!!info
    To better understand how this feature works, please read the [Create containers: Lifecycle Hooks](/features/creating_container/#lifecycle-hooks) documentation.

It leverages the `Executable` interface to represent the command and positional arguments to be executed in the container.

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the container is ready.

#### WithNetwork

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.27.0"><span class="tc-version">:material-tag: v0.27.0</span></a>

By default, the container is started in the default Docker network. If you want to use an already existing Docker network you created in your code, you can use the `network.WithNetwork(aliases []string, nw *testcontainers.DockerNetwork)` option, which receives an alias as parameter and your network, attaching the container to it, and setting the network alias for that network.

In the case you need to retrieve the network name, you can simply read it from the struct's `Name` field. E.g. `nw.Name`.

!!!warning
    This option is not checking whether the network exists or not. If you use a network that doesn't exist, the container will start in the default Docker network, as in the default behavior.

#### WithNewNetwork

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.27.0"><span class="tc-version">:material-tag: v0.27.0</span></a>

If you want to attach your containers to a throw-away network, you can use the `network.WithNewNetwork(ctx context.Context, aliases []string, opts ...network.NetworkCustomizer)` option, which receives an alias as parameter, creating the new network with a random name, attaching the container to it, and setting the network alias for that network.

In the case you need to retrieve the network name, you can use the `Networks(ctx)` method of the `Container` interface, right after it's running, which returns a slice of strings with the names of the networks where the container is attached.

#### Docker type modifiers

If you need an advanced configuration for the container, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](/features/creating_container.md#advanced-settings) documentation for more information.

#### Customising the ContainerRequest

This option will merge the customized request into the module's own `ContainerRequest`.

```go
container, err := Run(ctx, "postgres:13-alpine",
    /* Other module options */
    testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Cmd: []string{"-c", "log_statement=all"},
        },
    }),
)
```

The above example is updating the predefined command of the image, **appending** them to the module's command.

!!!info
    This can't be used to replace the command, only to append options.
