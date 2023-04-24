# Redis

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Redis.

## Adding this module to your project dependencies

Please run the following command to add the Redis module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/redis
```

## Usage example

<!--codeinclude-->
[Creating a Redis container](../../modules/redis/redis_test.go) inside_block:createRedisContainer
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
for Postgres. E.g. `testcontainers.WithImage("docker.io/redis:7")`.

<!--codeinclude-->
[Use a different image](../../modules/redis/redis_test.go) inside_block:withImage
<!--/codeinclude-->

#### Snapshotting

By default Redis saves snapshots of the dataset on disk, in a binary file called dump.rdb. You can configure Redis to have it save the dataset every N seconds if there are at least M changes in the dataset.

!!!tip
    Please check [Redis docs on persistence](https://redis.io/docs/management/persistence/#snapshotting) for more information.

<!--codeinclude-->
[Saving snapshots](../../modules/redis/redis_test.go) inside_block:withSnapshotting
<!--/codeinclude-->

#### Log Level

By default Redis saves snapshots of the dataset on disk, in a binary file called dump.rdb. You can configure Redis to have it save the dataset every N seconds if there are at least M changes in the dataset.

!!!tip
    Please check [Redis docs on logging](https://redis.io/docs/reference/modules/modules-api-ref/#redismodule_log) for more information.

<!--codeinclude-->
[Changing the log level](../../modules/redis/redis_test.go) inside_block:withLogLevel 
<!--/codeinclude-->

#### Redis configuration

In the case you have a custom config file for Redis, it's possible to copy that file into the container before it's started.

<!--codeinclude-->
[Include custom configuration file](../../modules/redis/redis_test.go) inside_block:withConfigFile
<!--/codeinclude-->

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
