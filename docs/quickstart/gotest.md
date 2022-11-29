_Testcontainers for Go_ plays well with the native `go test` framework.

The ideal use case is for integration or end to end tests. It helps you to spin
up and manage the dependencies life cycle via Docker.

Docker has to be available for this library to work.

## 1. Install

We use [gomod](https://blog.golang.org/using-go-modules) and you can get it installed via:

```
go get github.com/testcontainers/testcontainers-go
```

!!!warning

	Given the version includes the Compose dependency, and the Docker folks added [a replace directive until the upcoming Docker 22.06 release is out](https://github.com/docker/compose/issues/9946#issuecomment-1288923912),
	we were forced to add it too, causing consumers of _Testcontainers for Go_ to add the following replace directive to their `go.mod` files.
	We expect this to be removed in the next releases of _Testcontainers for Go_.

	```
	replace (
		github.com/docker/cli => github.com/docker/cli v20.10.3-0.20221013132413-1d6c6e2367e2+incompatible // 22.06 master branch
		github.com/docker/docker => github.com/docker/docker v20.10.3-0.20221013203545-33ab36d6b304+incompatible // 22.06 branch
		github.com/moby/buildkit => github.com/moby/buildkit v0.10.1-0.20220816171719-55ba9d14360a // same as buildx

		github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.2 // Can be removed on next bump of containerd to > 1.6.4

		// For k8s dependencies, we use a replace directive, to prevent them being
		// upgraded to the version specified in containerd, which is not relevant to the
		// version needed.
		// See https://github.com/docker/buildx/pull/948 for details.
		// https://github.com/docker/buildx/blob/v0.8.1/go.mod#L62-L64
		k8s.io/api => k8s.io/api v0.22.4
		k8s.io/apimachinery => k8s.io/apimachinery v0.22.4
		k8s.io/client-go => k8s.io/client-go v0.22.4
	)
	```

## 2. Spin up Redis

```go
import (
	"context"
	"testing"

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
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := redisC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err.Error())
		}
	}()
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
terminated function: `defer redisC.Terminate(ctx)`.

!!!tip

    Look at [features/garbage_collector](/features/garbage_collector/) to know another way to
    clean up resources.

## 3. Make your code to talk with the container

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

## 3. Run the test

You can run the test via `go test ./...`
