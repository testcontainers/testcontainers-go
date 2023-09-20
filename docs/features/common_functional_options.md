#### Wait Strategies

If you need to set a different wait strategy for the container, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Startup Commands

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Testcontainers exposes the `WithStartupCommand(e ...Executable)` option to run arbitrary commands in the container right after it's started.

!!!info
    To better understand how this feature works, please read the [Create containers: Lifecycle Hooks](/features/creating_container/#lifecycle-hooks) documentation.

It also exports an `Executable` interface, defining one single method: `AsCommand()`, which returns a slice of strings to represent the command and positional arguments to be executed in the container.

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the container is started.

#### Docker type modifiers

If you need an advanced configuration for the container, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](/features/creating_container.md#advanced-settings) documentation for more information.
