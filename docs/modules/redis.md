# Redis

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

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

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Redis module exposes one entrypoint function to create the container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*RedisContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "redis:7")`.

### Container Options

When starting the Redis container, you can pass options in a variadic way to configure it.

!!!tip
    You can find all the available configuration and environment variables for the Redis Docker image on [Docker Hub](https://hub.docker.com/_/redis).

#### WithSnapshotting

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

By default Redis saves snapshots of the dataset on disk, in a binary file called dump.rdb. You can configure Redis to have it save the dataset every `N` seconds if there are at least `M` changes in the dataset. E.g. `WithSnapshotting(10, 1)`.

!!!tip
    Please check [Redis docs on persistence](https://redis.io/docs/management/persistence/#snapshotting) for more information.

#### WithLogLevel

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

By default Redis produces a log message to the standard Redis log, the format accepts printf-alike specifiers, while level is a string describing the log level to use when emitting the log, and must be one of the following: `LogLevelDebug`, `LogLevelVerbose`, `LogLevelNotice`, `LogLevelWarning`. E.g. `WithLogLevel(LogLevelDebug)`. If the specified log level is invalid, verbose is used by default.

!!!tip
    Please check [Redis docs on logging](https://redis.io/docs/reference/modules/modules-api-ref/#redismodule_log) for more information.

#### WithConfigFile

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

In the case you have a custom config file for Redis, it's possible to copy that file into the container before it's started. E.g. `WithConfigFile(filepath.Join("testdata", "redis7.conf"))`.

#### WithTLS

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

In the case you want to enable TLS for the Redis container, you can use the `WithTLS()` option. This options enables TLS on the `6379/tcp` port and uses a secure URL (e.g. `rediss://host:port`).

!!!info
    In case you want to use Non-mutual TLS (i.e. client authentication is not required), you can customize the CMD arguments by using the `WithCmdArgs` option. E.g. `WithCmdArgs("--tls-auth-clients", "no")`.

The module automatically generates three certificates, a CA certificate, a client certificate and a Redis certificate. Please use the `TLSConfig()` container method to get the TLS configuration and use it to configure the Redis client. See more details in the [TLSConfig](#tlsconfig) section.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

#### ConnectionString

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

This method returns the connection string to connect to the Redis container, using the default `6379` port, and `redis` schema.

<!--codeinclude-->
[Get connection string](../../modules/redis/redis_test.go) inside_block:connectionString
<!--/codeinclude-->

If the container is started with TLS enabled, the connection string is `rediss://host:port`, using the `rediss` schema.

#### TLSConfig

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

This method returns the TLS configuration for the Redis container, nil if TLS is not enabled.

<!--codeinclude-->
[Get TLS config](../../modules/redis/redis_test.go) inside_block:tlsConfig
<!--/codeinclude-->

In the above example, the options are used to configure a Redis client with TLS enabled.
