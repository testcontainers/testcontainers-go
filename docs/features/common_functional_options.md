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

#### WithNetwork

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

By default, the container is started in the default Docker network. If you want to use a different Docker network, you can use the `WithNetwork(networkName string, alias string)` option, which receives the new network name and an alias as parameters, creating the new network, attaching the container to it, and setting the network alias for that network.

If the network already exists, _Testcontainers for Go_ won't create a new one, but it will attach the container to it and set the network alias.

#### Docker type modifiers

If you need an advanced configuration for the container, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](/features/creating_container.md#advanced-settings) documentation for more information.
