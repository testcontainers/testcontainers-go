# How to create a container

Testcontainers are a wrapper around the Docker daemon designed for tests. Anything you can run in Docker, you can spin
up with Testcontainers and integrate into your tests:

* NoSQL databases or other data stores (e.g. Redis, ElasticSearch, MongoDB)
* Web servers/proxies (e.g. NGINX, Apache)
* Log services (e.g. Logstash, Kibana)
* Other services developed by your team/organization which are already dockerized

## Run

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`testcontainers.Run` defines the container that should be run, similar to the `docker run` command.

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DockerContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

The following test creates an NGINX container on both the `bridge` (docker default
network) and the `foo` network and validates that it returns 200 for the status code.

It also demonstrates how to use `CleanupContainer`, that ensures that nginx container
is removed when the test ends even if the underlying container errored,
as well as the `CleanupNetwork` which does the same for networks.

The alternatives for these outside of tests as a `defer` are `TerminateContainer`
and `Network.Remove` which can be seen in the examples.

<!--codeinclude-->
[Creating a container](../../examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## GenericContainer

!!!warning
	`GenericContainer` is the old way to create a container, and we recommend using `Run` instead,
	as it could be deprecated in the future.

`testcontainers.GenericContainer` defines the container that should be run, similar to the `docker run` command.

The following test creates an NGINX container on both the `bridge` (docker default
network) and the `foo` network and validates that it returns 200 for the status code.

It also demonstrates how to use `CleanupContainer` ensures that nginx container
is removed when the test ends even if the underlying `GenericContainer` errored
as well as the `CleanupNetwork` which does the same for networks.

The alternatives for these outside of tests as a `defer` are `TerminateContainer`
and `Network.Remove` which can be seen in the examples.

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type nginxContainer struct {
	testcontainers.Container
	URI string
}


func setupNginx(ctx context.Context, networkName string) (*nginxContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		Networks:     []string{"bridge", networkName},
		WaitingFor:   wait.ForHTTP("/"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	var nginxC *nginxContainer
	if container != nil {
		nginxC = &nginxContainer{Container: container}
	}
	if err != nil {
		return nginxC, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nginxC, err
	}

	mappedPort, err := container.MappedPort(ctx, "80")
	if err != nil {
		return nginxC, err
	}

	nginxC.URI = fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	return nginxC, nil
}

func TestIntegrationNginxLatestReturn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	nw, err := network.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, nw)

	nginxC, err := setupNginx(ctx, nw.Name)
	testcontainers.CleanupContainer(t, nginxC)
	require.NoError(t, err)

	resp, err := http.Get(nginxC.URI)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
```




### Lifecycle hooks

_Testcontainers for Go_ allows you to define your own lifecycle hooks for better control over your containers. You just need to define functions that return an error and receive the Go context as first argument, and a `ContainerRequest` for the `Creating` hook, and a `Container` for the rest of them as second argument.

You'll be able to pass multiple lifecycle hooks at the `ContainerRequest` as an array of `testcontainers.ContainerLifecycleHooks`. The `testcontainers.ContainerLifecycleHooks` struct defines the following lifecycle hooks, each of them backed by an array of functions representing the hooks:

* `PreBuilds` - hooks that are executed before the image is built. This hook is only available when creating a container from a Dockerfile
* `PostBuilds` - hooks that are executed after the image is built. This hook is only available when creating a container from a Dockerfile
* `PreCreates` - hooks that are executed before the container is created
* `PostCreates` - hooks that are executed after the container is created
* `PreStarts` - hooks that are executed before the container is started
* `PostStarts` - hooks that are executed after the container is started
* `PostReadies` - hooks that are executed after the container is ready
* `PreStops` - hooks that are executed before the container is stopped
* `PostStops` - hooks that are executed after the container is stopped
* `PreTerminates` - hooks that are executed before the container is terminated
* `PostTerminates` - hooks that are executed after the container is terminated

_Testcontainers for Go_ defines some default lifecycle hooks that are always executed in a specific order with respect to the user-defined hooks. The order of execution is the following:

1. default `pre` hooks.
2. user-defined `pre` hooks.
3. user-defined `post` hooks.
4. default `post` hooks.

Inside each group, the hooks will be executed in the order they were defined.

!!!info
	The default hooks are for logging (applied to all hooks), customising the Docker config (applied to the pre-create hook), copying files in to the container (applied to the post-create hook), adding log consumers (applied to the post-start and pre-terminate hooks), and running the wait strategies as a readiness check (applied to the post-start hook).

It's important to notice that the `Readiness` of a container is defined by the wait strategies defined for the container. **This hook will be executed right after the `PostStarts` hook**. If you want to add your own readiness checks, you can do it by adding a `PostReadies` hook to the container request, which will execute your own readiness check after the default ones. That said, the `PostStarts` hooks don't warrant that the container is ready, so you should not rely on that.

In the following example, we are going to create a container using all the lifecycle hooks, all of them printing a message when any of the lifecycle hooks is called:

<!--codeinclude-->
[Extending container with lifecycle hooks](../../lifecycle_test.go) inside_block:reqWithLifecycleHooks
<!--/codeinclude-->

#### Default Logging Hook

_Testcontainers for Go_ comes with a default logging hook that will print a log message for each container lifecycle event, using the default logger. You can add your own logger by passing the `testcontainers.DefaultLoggingHook` option to the `ContainerRequest`, passing a reference to your preferred logger:

<!--codeinclude-->
[Use a custom logger for container hooks](../../lifecycle_test.go) inside_block:reqWithDefaultLoggingHook
[Custom Logger implementation](../../lifecycle_test.go) inside_block:customLoggerImplementation
<!--/codeinclude-->

### Advanced Settings

The aforementioned `GenericContainer` function and the `ContainerRequest` struct represent a straightforward manner to configure the containers, but you could need to create your containers with more advance settings regarding the config, host config and endpoint settings Docker types. For those more advance settings, _Testcontainers for Go_ offers a way to fully customize the container request and those internal Docker types. These customisations, called _modifiers_, will be applied just before the internal call to the Docker client to create the container.

<!--codeinclude-->
[Using modifiers](../../lifecycle_test.go) inside_block:reqWithModifiers
<!--/codeinclude-->

!!!warning
	The only special case where the modifiers are not applied last, is when there are no exposed ports in the container request and the container does not use a network mode from a container (e.g. `req.NetworkMode = container.NetworkMode("container:$CONTAINER_ID")`). In that case, _Testcontainers for Go_ will extract the ports from the underlying Docker image and export them.

## Reusable container

With `Reuse` option you can reuse an existing container. Reusing will work only if you pass an
existing container name via 'req.Name' field. If the name is not in a list of existing containers,
the function will create a new generic container. If `Reuse` is true and `Name` is empty, you will get error.

The following test creates an NGINX container, adds a file into it and then reuses the container again for checking the file:
```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	reusableContainerName = "my_test_reusable_container"
)

func main() {
	ctx := context.Background()

	n1, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
			Name:         reusableContainerName,
		},
		Started: true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(n1); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Print(err)
		return
	}

	copiedFileName := "hello_copy.sh"
	err = n1.CopyFileToContainer(ctx, "./testdata/hello.sh", "/"+copiedFileName, 700)

	if err != nil {
		log.Print(err)
		return
	}

	n2, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
			Name:         reusableContainerName,
		},
		Started: true,
		Reuse:   true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(n2); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Print(err)
		return
	}

	c, _, err := n2.Exec(ctx, []string{"bash", copiedFileName})
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(c)
}
```

## Parallel running

`testcontainers.ParallelContainers` - defines the containers that should be run in parallel mode.

The following test creates two NGINX containers in parallel:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
)

func main() {
	ctx := context.Background()

	requests := testcontainers.ParallelContainerRequest{
		{
			ContainerRequest: testcontainers.ContainerRequest{

				Image: "nginx",
				ExposedPorts: []string{
					"10080/tcp",
				},
			},
			Started: true,
		},
		{
			ContainerRequest: testcontainers.ContainerRequest{

				Image: "nginx",
				ExposedPorts: []string{
					"10081/tcp",
				},
			},
			Started: true,
		},
	}

	res, err := testcontainers.ParallelContainers(ctx, requests, testcontainers.ParallelContainersOptions{})
	for _, c := range res {
		c := c
		defer func() {
			if err := testcontainers.TerminateContainer(c); err != nil {
				log.Printf("failed to terminate container: %s", c)
			}
		}()
	}

	if err != nil {
		e, ok := err.(testcontainers.ParallelContainersError)
		if !ok {
			log.Printf("unknown error: %v", err)
			return
		}

		for _, pe := range e.Errors {
			fmt.Println(pe.Request, pe.Error)
		}
		return
	}
}
```
