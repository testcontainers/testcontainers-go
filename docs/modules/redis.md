# Redis

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

## Introduction

The Testcontainers module for Redis.

## Adding this module to your project dependencies

Please run the following command to add the Redis module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/redis
```

## Usage example

<!--codeinclude-->
[Creating a Redis container](../../modules/redis/examples_test.go) inside_block:runRedisContainer
<!--/codeinclude-->

## Module Reference

The Redis module exposes one entrypoint function to create the containerr, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RedisContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Redis container, you can pass options in a variadic way to configure it.

!!!tip
    You can find all the available configuration and environment variables for the Redis Docker image on [Docker Hub](https://hub.docker.com/_/redis).

#### Image

If you need to set a different Redis Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Redis. E.g. `testcontainers.WithImage("docker.io/redis:7")`.

#### Wait Strategies

If you need to set a different wait strategy for Redis, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for Redis.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for Redis, you can leverage the following Docker type modifiers:

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

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the Redis container is started.

#### Snapshotting

By default Redis saves snapshots of the dataset on disk, in a binary file called dump.rdb. You can configure Redis to have it save the dataset every `N` seconds if there are at least `M` changes in the dataset. E.g. `WithSnapshotting(10, 1)`.

!!!tip
    Please check [Redis docs on persistence](https://redis.io/docs/management/persistence/#snapshotting) for more information.

#### Log Level

By default Redis saves snapshots of the dataset on disk, in a binary file called dump.rdb. You can configure Redis to have it save the dataset every N seconds if there are at least M changes in the dataset. E.g. `WithLogLevel(LogLevelDebug)`.

!!!tip
    Please check [Redis docs on logging](https://redis.io/docs/reference/modules/modules-api-ref/#redismodule_log) for more information.

#### Redis configuration

In the case you have a custom config file for Redis, it's possible to copy that file into the container before it's started. E.g. `WithConfigFile(filepath.Join("testdata", "redis7.conf"))`.

### Container Methods

#### ConnectionString

This method returns the connection string to connect to the Redis container, using the default `6379` port.

<!--codeinclude-->
[Get connection string](../../modules/redis/redis_test.go) inside_block:connectionString
<!--/codeinclude-->

### Redis variants

It's possible to use the Redis container with Redis-Stack. You simply need to update the image name.

<!--codeinclude-->
[Image for Redis-Stack](../../modules/redis/redis_test.go) inside_block:redisStackImage
<!--/codeinclude-->

<!--codeinclude-->
[Image for Redis-Stack Server](../../modules/redis/redis_test.go) inside_block:redisStackServerImage
<!--/codeinclude-->
