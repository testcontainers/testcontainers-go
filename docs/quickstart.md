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
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	testcontainers.CleanupContainer(t, redisC)
	require.NoError(t, err)
}
```

The `testcontainers.ContainerRequest` describes how the Docker container will
look.

* `Image` is the Docker image the container starts from.
* `ExposedPorts` lists the ports to be exposed from the container.
* `WaitingFor` is a field you can use to validate when a container is ready. It
  is important to get this set because it helps to know when the container is
  ready to receive any traffic. In this case, we check for the logs we know come
  from Redis, telling us that it is ready to accept requests.

When you use `ExposedPorts` you have to imagine yourself using `docker run -p
<port>`.  When you do so, `dockerd` maps the selected `<port>` from inside the
container to a random one available on your host.

In the previous example, we expose `6379` for `tcp` traffic to the outside. This
allows Redis to be reachable from your code that runs outside the container, but
it also makes parallelization possible because if you add `t.Parallel` to your
tests, and each of them starts a Redis container each of them will be exposed on a
different random port.

`testcontainers.GenericContainer` creates the container. In this example we are
using `Started: true`. It means that the container function will wait for the
container to be up and running. If you set the `Start` value to `false` it won't
start, leaving to you the decision about when to start it.

All the containers must be removed at some point, otherwise they will run until
the host is overloaded. One of the ways we have to clean up is by deferring the
terminated function: `defer testcontainers.TerminateContainer(redisC)` which
automatically handles nil container so is safe to use even in the error case.

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
