# Garbage Collector

Typically, an integration test creates one or more containers. This can mean a
lot of containers running by the time everything is done. We need to have a way
to clean up after ourselves to keep our machines running smoothly.

Containers can be unused because:

1. Test is over and the container is not needed anymore.
2. Test failed, we do not need that container anymore because next build will
   create new containers.

## Terminate function

As we saw previously there are at least two ways to remove unused containers.
The primary method is to use the `Terminate(context.Context)` function that is
available when a container is created. Use `defer` to ensure that it is called
on test completion.

The `Terminate` function can be customised with termination options to determine how a container is removed: termination timeout, and the ability to remove container volumes are supported at the moment. You can build the default options using the `testcontainers.NewTerminationOptions` function.

#### NewTerminateOptions

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.35.0"><span class="tc-version">:material-tag: v0.35.0</span></a>

If you want to attach option to container termination, you can use the `testcontainers.NewTerminateOptions(ctx context.Context, opts ...TerminateOption) *TerminateOptions` option, which receives a TerminateOption as parameter, creating custom termination options to be passed on the container termination.

##### Terminate Options

###### [StopContext](../../cleanup.go)
Sets the context for the Container termination.

- **Function**: `StopContext(ctx context.Context) TerminateOption`
- **Default**: The context passed in `Terminate()`
- **Usage**:
```go
err := container.Terminate(ctx,StopContext(context.Background()))
```

###### [StopTimeout](../../cleanup.go)
Sets the timeout for stopping the Container.

- **Function**: ` StopTimeout(timeout time.Duration) TerminateOption`
- **Default**:  10 seconds
- **Usage**:
```go
err := container.Terminate(ctx, StopTimeout(20 * time.Second))
```

###### [RemoveVolumes](../../cleanup.go)
Sets the volumes to be removed during Container termination.

- **Function**: ` RemoveVolumes(volumes ...string) TerminateOption`
- **Default**:  Empty (no volumes removed)
- **Usage**:
```go
err := container.Terminate(ctx, RemoveVolumes("vol1", "vol2"))
```


!!!tip

    Remember to `defer` as soon as possible so you won't forget. The best time
    is as soon as you call `testcontainers.GenericContainer` but remember to
    check for the `err` first.

## Ryuk

[Ryuk](https://github.com/testcontainers/moby-ryuk) (also referred to as
`Reaper` in this package) removes containers/networks/volumes created by
_Testcontainers for Go_ after a specified delay. It is a project developed by the
Testcontainers organization and is used across the board for many of the
different language implementations.

When you run one test, you will see an additional container called `ryuk`
alongside all of the containers that were specified in your test. It relies on
container labels to determine which resources were created by the package
to determine the entities that are safe to remove. If a container is running
for more than 10 seconds, it will be killed.

!!!warning

    This feature can be disabled in two different manners, but it can cause **unexpected behavior** in your environment:

    1. adding `ryuk.disabled=true` to the `.testcontainers.properties` file.
    2. setting the `TESTCONTAINERS_RYUK_DISABLED=true` environment variable. This manner takes precedence over the properties file.

    We recommend using it only for Continuous Integration services that have their
    own mechanism to clean up resources.

Even if you do not call Terminate, Ryuk ensures that the environment will be
kept clean and even cleans itself when there is nothing left to do.
