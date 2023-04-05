# How to create a container

Testcontainers are a wrapper around the Docker daemon designed for tests. Anything you can run in Docker, you can spin
up with Testcontainers and integrate into your tests:
* NoSQL databases or other data stores (e.g. Redis, ElasticSearch, MongoDB)
* Web servers/proxies (e.g. NGINX, Apache)
* Log services (e.g. Logstash, Kibana)
* Other services developed by your team/organization which are already dockerized

## GenericContainer

`testcontainers.GenericContainer` defines the container that should be run, similar to the `docker run` command.

The following test creates an NGINX container and validates that it returns 200 for the status code:

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


func setupNginx(ctx context.Context) (*nginxContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "80")
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	return &nginxContainer{Container: container, URI: uri}, nil
}

func TestIntegrationNginxLatestReturn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	nginxC, err := setupNginx(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := nginxC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	resp, err := http.Get(nginxC.URI)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}
```

### Lifecycle hooks

_Testcontainers for Go_ allows you to define your own lifecycle hooks for better control over your containers. You just need to define functions that return an error and receive the Go context as first argument, and a `ContainerRequest` for the `Creating` hook, and a `Container` for the rest of them as second argument.

You'll be able to pass multiple lifecycle hooks at the `ContainerRequest` as an array of `testcontainers.ContainerLifecycleHooks`, which will be processed one by one in the order they are passed.

The `testcontainers.ContainerLifecycleHooks` struct defines the following lifecycle hooks, each of them backed by an array of functions representing the hooks:

* `PreCreates` - hooks that are executed before the container is created
* `PostCreates` - hooks that are executed after the container is created
* `PreStarts` - hooks that are executed before the container is started
* `PostStarts` - hooks that are executed after the container is started
* `PreStops` - hooks that are executed before the container is stopped
* `PostStops` - hooks that are executed after the container is stopped
* `PreTerminates` - hooks that are executed before the container is terminated
* `PostTerminates` - hooks that are executed after the container is terminated

In the following example, we are going to create a container using all the lifecycle hooks, all of them printing a message when any of the lifecycle hooks is called:

<!--codeinclude-->
[Extending container with lifecycle hooks](../../lifecycle_test.go) inside_block:reqWithLifecycleHooks
<!--/codeinclude-->

#### Default Logging Hook

_Testcontainers for Go_ comes with a default logging hook that will print a log message for each container lifecycle event. You can enable it by passing the `testcontainers.DefaultLoggingHook` option to the `ContainerRequest`, passing a reference to the container logger like this:

<!--codeinclude-->
[Extending container with life cycle hooks](../../lifecycle_test.go) inside_block:reqWithDefaultLogginHook
<!--/codeinclude-->

### Advanced Settings

The aforementioned `GenericContainer` function and the `ContainerRequest` struct represent a straightforward manner to configure the containers, but you could need to create your containers with more advance settings regarding the config, host config and endpoint settings Docker types. For those more advance settings, _Testcontainers for Go_ offers a way to fully customize the container request and those internal Docker types. These customisations, called _modifiers_, will be applied just before the internal call to the Docker client to create the container.

<!--codeinclude-->
[Using modifiers](../../lifecycle_test.go) inside_block:reqWithModifiers
<!--/codeinclude-->

!!!warning
	The only special case where the modifiers are not applied last, is when there are no exposed ports in the container request and the container does not use a network mode from a container (e.g. `req.NetworkMode = container.NetworkMode("container:$CONTAINER_ID")`). In that case, _Testcontainers for Go_ will extract the ports from the underliying Docker image and export them.

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
)

const (
	reusableContainerName = "my_test_reusable_container"
)

ctx := context.Background()

n1, err := GenericContainer(ctx, GenericContainerRequest{
	ContainerRequest: ContainerRequest{
		Image:        "nginx:1.17.6",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForListeningPort("80/tcp"),
		Name:         reusableContainerName,
	},
	Started: true,
})
if err != nil {
	log.Fatal(err)
}
defer n1.Terminate(ctx)

copiedFileName := "hello_copy.sh"
err = n1.CopyFileToContainer(ctx, "./testdata/hello.sh", "/"+copiedFileName, 700)

if err != nil {
	log.Fatal(err)
}

n2, err := GenericContainer(ctx, GenericContainerRequest{
	ContainerRequest: ContainerRequest{
		Image:        "nginx:1.17.6",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForListeningPort("80/tcp"),
		Name:         reusableContainerName,
    },
	Started: true,
	Reuse: true,
})
if err != nil {
	log.Fatal(err)
}

c, _, err := n2.Exec(ctx, []string{"bash", copiedFileName})
if err != nil {
	log.Fatal(err)
}
fmt.Println(c)
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
	if err != nil {
		e, ok := err.(testcontainers.ParallelContainersError)
		if !ok {
			log.Fatalf("unknown error: %v", err)
		}

		for _, pe := range e.Errors {
			fmt.Println(pe.Request, pe.Error)
		}
		return
	}

	for _, c := range res {
		c := c
		defer func() {
			if err := c.Terminate(ctx); err != nil {
				log.Fatalf("failed to terminate container: %s", c)
			}
		}()
	}
}
```
