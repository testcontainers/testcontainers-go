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

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Redis module exposes one entrypoint function to create the container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*RedisContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Redis container, you can pass options in a variadic way to configure it.

!!!tip
    You can find all the available configuration and environment variables for the Redis Docker image on [Docker Hub](https://hub.docker.com/_/redis).

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "redis:7")`.

{% include "../features/common_functional_options.md" %}

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
