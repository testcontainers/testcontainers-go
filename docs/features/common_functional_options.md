#### Basic Options

##### WithExposedPorts

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

If you need to expose additional ports from the container, you can use `testcontainers.WithExposedPorts`. For example:

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithExposedPorts("8080/tcp", "9090/tcp"))
```

##### WithEnv

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

If you need to either pass additional environment variables to a container or override them, you can use `testcontainers.WithEnv` for example:

```golang
ctr, err = mymodule.Run(ctx, "docker.io/myservice:1.2.3", testcontainers.WithEnv(map[string]string{"FOO": "BAR"}))
```

##### WithWaitStrategy

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

If you need to set a different wait strategy for the container, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy.

!!!info
    The default deadline for the wait strategy is 60 seconds.

##### WithWaitStrategyAndDeadline

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

##### WithAdditionalWaitStrategy

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you need to add a wait strategy to the existing wait strategy, you can use `testcontainers.WithAdditionalWaitStrategy`.

!!!info
    The default deadline for the wait strategy is 60 seconds.

##### WithAdditionalWaitStrategyAndDeadline

- - Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

At the same time, it's possible to add a wait strategy and a custom deadline with `testcontainers.WithAdditionalWaitStrategyAndDeadline`.

##### WithEntrypoint

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

If you need to completely replace the container's entrypoint, you can use `testcontainers.WithEntrypoint`. For example:

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithEntrypoint("/bin/sh", "-c", "echo hello"))
```

##### WithEntrypointArgs

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

If you need to append commands to the container's entrypoint, you can use `testcontainers.WithEntrypointArgs`. For example:

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithEntrypointArgs("echo", "hello"))
```

##### WithCmd

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

If you need to completely replace the container's command, you can use `testcontainers.WithCmd`. For example:

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithCmd("echo", "hello"))
```

##### WithCmdArgs

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

If you need to append commands to the container's command, you can use `testcontainers.WithCmdArgs`. For example:

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithCmdArgs("echo", "hello"))
```

##### WithLabels

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

If you need to add Docker labels to the container, you can use `testcontainers.WithLabels`. For example:

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithLabels(map[string]string{
        "environment": "testing",
        "project":     "myapp",
    }))
```

#### Lifecycle Options

##### WithLifecycleHooks

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you need to set the lifecycle hooks for the container, you can use `testcontainers.WithLifecycleHooks`, which replaces the existing lifecycle hooks with the new ones.

##### WithAdditionalLifecycleHooks

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

You can also use `testcontainers.WithAdditionalLifecycleHooks`, which appends the new lifecycle hooks to the existing ones.

##### WithStartupCommand

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.25.0"><span class="tc-version">:material-tag: v0.25.0</span></a>

Testcontainers exposes the `WithStartupCommand(e ...Executable)` option to run arbitrary commands in the container right after it's started.

!!!info
    To better understand how this feature works, please read the [Create containers: Lifecycle Hooks](/features/creating_container/#lifecycle-hooks) documentation.

It also exports an `Executable` interface, defining the following methods:

- `AsCommand()`, which returns a slice of strings to represent the command and positional arguments to be executed in the container;
- `Options()`, which returns the slice of functional options with the Docker's ExecConfigs used to create the command in the container (the working directory, environment variables, user executing the command, etc) and the possible output format (Multiplexed).

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the container is started.

##### WithAfterReadyCommand

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

Testcontainers exposes the `WithAfterReadyCommand(e ...Executable)` option to run arbitrary commands in the container right after it's ready, which happens when the defined wait strategies have finished with success.

!!!info
    To better understand how this feature works, please read the [Create containers: Lifecycle Hooks](/features/creating_container/#lifecycle-hooks) documentation.

It leverages the `Executable` interface to represent the command and positional arguments to be executed in the container.

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the container is ready.

#### Files & Mounts Options

##### WithFiles

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

If you need to copy files into the container, you can use `testcontainers.WithFiles`. For example:

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithFiles([]testcontainers.ContainerFile{
        {
            HostFilePath:      "/path/to/local/file.txt",
            ContainerFilePath: "/container/file.txt",
            FileMode:          0o644,
        },
    }))
```

This option allows you to copy files from the host into the container at creation time.

##### WithMounts

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

If you need to add volume mounts to the container, you can use `testcontainers.WithMounts`. For example:

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithMounts([]testcontainers.ContainerMount{
        {
            Source: testcontainers.GenericVolumeMountSource{Name: "appdata"},
            Target: "/app/data",
        },
    }))
```

##### WithTmpfs

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

If you need to add tmpfs mounts to the container, you can use `testcontainers.WithTmpfs`. For example:

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithTmpfs(map[string]string{
        "/tmp": "size=100m",
        "/run": "size=100m",
    }))
```

##### WithImageMount

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

Since Docker v28, it's possible to mount an image to a container, passing the source image name, the relative subpath to mount in that image, and the mount point in the target container.

This option validates that the subpath is a relative path, raising an error otherwise.

<!--codeinclude-->
[Image Mount](../../modules/ollama/examples_test.go) inside_block:mountImage
<!--/codeinclude-->

In the code above, which mounts the directory in which Ollama models are stored, the `targetImage` is the name of the image containing the models (an Ollama image where the models are already pulled).

!!!warning
    Using this option fails the creation of the container if the underlying container runtime does not support the `image mount` feature.

#### Build Options

##### WithDockerfile

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

Testcontainers exposes the `testcontainers.WithDockerfile` option to build a container from a Dockerfile.
The functional option receives a `testcontainers.FromDockerfile` struct that is applied to the container request before starting the container. As a result, the container is built and started in one go.

```golang
df := testcontainers.FromDockerfile{
	Context:    ".",
	Dockerfile: "Dockerfile",
	Repo:       "testcontainers",
	Tag:        "latest",
	BuildArgs:  map[string]*string{"ARG1": nil, "ARG2": nil},
}   

ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", testcontainers.WithDockerfile(df))
```

#### Logging Options

##### WithLogConsumers

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

If you need to consume the logs of the container, you can use `testcontainers.WithLogConsumers` with a valid log consumer. An example of a log consumer is the following:

```golang
type TestLogConsumer struct {
	Msgs []string
}

func (g *TestLogConsumer) Accept(l Log) {
	g.Msgs = append(g.Msgs, string(l.Content))
}
```

##### WithLogConsumerConfig

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you need to set the log consumer config for the container, you can use `testcontainers.WithLogConsumerConfig`. This option completely replaces the existing log consumer config, including the log consumers and the log production options.

##### WithLogger

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

If you need to either pass logger to a container, you can use `testcontainers.WithLogger`.

!!!info
	Consider calling this before other "With" functions as these may generate logs.

In this example we also use the testcontainers-go `log.TestLogger`, which writes to the passed in `testing.TB` using `Logf`.
The result is that we capture all logging from the container into the test context meaning its
hidden behind `go test -v` and is associated with the relevant test, providing the user with
useful context instead of appearing out of band.

```golang
func TestHandler(t *testing.T) {
    logger := log.TestLogger(t)
    ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", testcontainers.WithLogger(logger))
    CleanupContainer(t, ctr)
    require.NoError(t, err)
    // Do something with container.
}
```

Please read the [Following Container Logs](/features/follow_logs) documentation for more information about creating log consumers.

#### Image Options

##### WithAlwaysPull

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you need to pull the image before starting the container, you can use `testcontainers.WithAlwaysPull()`.

##### WithImageSubstitutors

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.26.0"><span class="tc-version">:material-tag: v0.26.0</span></a>

In more locked down / secured environments, it can be problematic to pull images from Docker Hub and run them without additional precautions.

An image name substitutor converts a Docker image name, as may be specified in code, to an alternative name. This is intended to provide a way to override image names, for example to enforce pulling of images from a private registry.

_Testcontainers for Go_ exposes an interface to perform this operation: `ImageSubstitutor`, and a No-operation implementation to be used as reference for custom implementations:

<!--codeinclude-->
[Image Substitutor Interface](../../options.go) inside_block:imageSubstitutor
[Noop Image Substitutor](../../container_test.go) inside_block:noopImageSubstitutor
<!--/codeinclude-->

Using the `WithImageSubstitutors` options, you could define your own substitutions to the container images. E.g. adding a prefix to the images so that they can be pulled from a Docker registry other than Docker Hub. This is the usual mechanism for using Docker image proxies, caches, etc.

##### WithImagePlatform

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you need to set the platform for a container, you can use `testcontainers.WithImagePlatform(platform string)`.

#### Networking Options

##### WithNetwork

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.27.0"><span class="tc-version">:material-tag: v0.27.0</span></a>

By default, the container is started in the default Docker network. If you want to use an already existing Docker network you created in your code, you can use the `network.WithNetwork(aliases []string, nw *testcontainers.DockerNetwork)` option, which receives an alias as parameter and your network, attaching the container to it, and setting the network alias for that network.

In the case you need to retrieve the network name, you can simply read it from the struct's `Name` field. E.g. `nw.Name`.

!!!warning
    This option is not checking whether the network exists or not. If you use a network that doesn't exist, the container will start in the default Docker network, as in the default behavior.

##### WithNetworkByName

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you want to attach your containers to an already existing Docker network by its name, you can use the `network.WithNetworkName(aliases []string, networkName string)` option, which receives an alias as parameter and the network name, attaching the container to it, and setting the network alias for that network.

!!!warning
    In case the network name is `bridge`, no aliases are set. This is because network-scoped alias is supported only for containers in user defined networks.

##### WithBridgeNetwork

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you want to attach your containers to the `bridge` network, you can use the `network.WithBridgeNetwork()` option.

!!!warning
    The `bridge` network is the default network for Docker. It's not a user defined network, so it doesn't support network-scoped aliases.

##### WithNewNetwork

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.27.0"><span class="tc-version">:material-tag: v0.27.0</span></a>

If you want to attach your containers to a throw-away network, you can use the `network.WithNewNetwork(ctx context.Context, aliases []string, opts ...network.NetworkCustomizer)` option, which receives an alias as parameter, creating the new network with a random name, attaching the container to it, and setting the network alias for that network.

In the case you need to retrieve the network name, you can use the `Networks(ctx)` method of the `Container` interface, right after it's running, which returns a slice of strings with the names of the networks where the container is attached.

#### Advanced Options

##### WithHostPortAccess

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.31.0"><span class="tc-version">:material-tag: v0.31.0</span></a>

If you need to access a port that is already running in the host, you can use `testcontainers.WithHostPortAccess` for example:

```golang
ctr, err = mymodule.Run(ctx, "docker.io/myservice:1.2.3", testcontainers.WithHostPortAccess(8080))
```

To understand more about this feature, please read the [Exposing host ports to the container](/features/networking/#exposing-host-ports-to-the-container) documentation.

##### WithConfigModifier

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

If you need an advanced configuration for the container, modifying the container's configuration, you can use the `testcontainers.WithConfigModifier` option, which gives access to the underlying Docker's Config type.

##### WithHostConfigModifier

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

If you need an advanced configuration for the container, modifying the container's host configuration, you can use the `testcontainers.WithHostConfigModifier` option, which gives access to the underlying Docker's HostConfig type.

##### WithEndpointSettingsModifier

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

If you need an advanced configuration for the container, modifying the container's endpoint settings, you can use the `testcontainers.WithEndpointSettingsModifier` option, which gives access to the underlying Docker's EndpointSettings type.

##### CustomizeRequest

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

This option will merge the customized request into the module's own `ContainerRequest`.

```go
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3",
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

##### WithName

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you need to set the name of the container, you can use the `testcontainers.WithName` option.

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithName("my-container-name"),
)
```

!!!warning
    This option is not checking whether the container name is already in use. If you use a name that is already in use, an error is returned.
    At the same time, we discourage using this option as it might lead to unexpected behavior, but we understand that in some cases it might be useful.

##### WithNoStart

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you need to prevent the container from being started after creation, you can use the `testcontainers.WithNoStart` option.

#### Experimental Options

##### WithReuseByName

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

This option marks a container to be reused if it exists or create a new one if it doesn't.
With the current implementation, the container name must be provided to identify the container to be reused.

```golang
ctr, err := mymodule.Run(ctx, "docker.io/myservice:1.2.3", 
    testcontainers.WithReuseByName("my-container-name"),
)
```

!!!warning
    Reusing a container is experimental and the API is subject to change for a more robust implementation that is not based on container names.
