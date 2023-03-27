# Redis

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

The Redis module exposes one entrypoint function to create the container, and this function receives one parameter:

```golang
func StartContainer(ctx context.Context) (*RedisContainer, error) {
```

- `context.Context`, the Go context.

### Container Methods

#### ConnectionString

This method returns the connection string to connect to the Redis container, using the default `6379` port.

<!--codeinclude-->
[Get connection string](../../modules/redis/redis_test.go) inside_block:connectionString
<!--/codeinclude-->
