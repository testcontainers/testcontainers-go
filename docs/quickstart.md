_Testcontainers for Go_ plays well with the native `go test` framework.

The ideal use case is for integration or end to end tests. It helps you to spin
up and manage the dependencies life cycle via Docker.

## 1. System requirements

Please read the [system requirements](../system_requirements/) page before you start.

## 2. Install _Testcontainers for Go_

We use [go mod](https://blog.golang.org/using-go-modules) and you can get it installed via:

```
go get github.com/testcontainers/testcontainers-go
```

## 3. Spin up Redis

```go
import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestWithRedis(t *testing.T) {
	ctx := context.Background()
	redisC, err := testcontainers.Run(
		ctx, "redis:latest",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp"),
			wait.ForLog("Ready to accept connections"),
		),
	)
	testcontainers.CleanupContainer(t, redisC)
	require.NoError(t, err)
}
```

The `testcontainers.Run` function receives the image name and a list of options with how the Docker container will
look.

* `WithExposedPorts` is an option that lists the ports to be exposed from the container.
* `WithWaitStrategy` is an option that you can use to validate when a container is ready. It
  is important to get this set because it helps to know when the container is
  ready to receive any traffic. In this case, we check for the logs we know come
  from Redis, telling us that it is ready to accept requests.

When you use `WithExposedPorts` you have to imagine yourself using `docker run -p
<port>`.  When you do so, `dockerd` maps the selected `<port>` from inside the
container to a random one available on your host.

In the previous example, we expose `6379` for `tcp` traffic to the outside. This
allows Redis to be reachable from your code that runs outside the container, but
it also makes parallelization possible because if you add `t.Parallel()` to your
tests, and each of them starts a Redis container each of them will be exposed on a
different random port.

`testcontainers.Run` creates and starts the container. In this example we are
not using the `WithNoStart()` option. It means that the container function will wait for the
container to be up and running. If you pass the `WithNoStart()` option, it won't
start, leaving to you the decision about when to start it.	

All the containers must be removed at some point, otherwise they will run until
Ryuk the resource reaper terminates them, or when the host is overloaded.
One of the ways we have to clean up the container immediately is by using the `testing`
package from the standard library along with the `CleanupContainer` function:
`testcontainers.CleanupContainer(t, redisC)`. This function
automatically handles nil container so it can be used before any error check.

!!!tip

    Look at [features/garbage_collector](/features/garbage_collector/) to know another way to
    clean up resources.

## 4. Make your code to talk with the container

This is just an example, but usually Go applications that rely on Redis are
using the [redis-go](https://github.com/go-redis/redis) client. This code gets
the endpoint from the container we just started, and it configures the client.

```go
endpoint, err := redisC.Endpoint(ctx, "")
if err != nil {
    t.Error(err)
}

client := redis.NewClient(&redis.Options{
    Addr: endpoint,
})

_ = client
```

We expose only one port, so the `Endpoint` does not need a second argument set.

!!!tip

    If you expose more than one port you can specify the one you need as a second
    argument.

In this case it returns: `localhost:<mappedportfor-6379>`.

## 5. Run the test

You can run the test via `go test ./...`

## 6. Want to go deeper with Redis?

You can find a more elaborated Redis example in our examples section. Please check it out [here](./modules/redis.md).
