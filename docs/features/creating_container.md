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

You'll be able to pass multiple lifecycle hooks at the `ContainerRequest` as an array of `testcontainers.ContainerLifecycleHooks`. The `testcontainers.ContainerLifecycleHooks` struct defines the following lifecycle hooks, each of them backed by an array of functions representing the hooks:

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
[Use a custom logger for container hooks](../../lifecycle_test.go) inside_block:reqWithDefaultLogginHook
[Custom Logger implementation](../../lifecycle_test.go) inside_block:customLoggerImplementation
<!--/codeinclude-->

### Advanced Settings

The aforementioned `GenericContainer` function and the `ContainerRequest` struct represent a straightforward manner to configure the containers, but you could need to create your containers with more advance settings regarding the config, host config and endpoint settings Docker types. For those more advance settings, _Testcontainers for Go_ offers a way to fully customize the container request and those internal Docker types. These customisations, called _modifiers_, will be applied just before the internal call to the Docker client to create the container.

<!--codeinclude-->
[Using modifiers](../../lifecycle_test.go) inside_block:reqWithModifiers
<!--/codeinclude-->

!!!warning
	The only special case where the modifiers are not applied last, is when there are no exposed ports in the container request and the container does not use a network mode from a container (e.g. `req.NetworkMode = container.NetworkMode("container:$CONTAINER_ID")`). In that case, _Testcontainers for Go_ will extract the ports from the underliying Docker image and export them.

## Reusable container

!!!warning
	Reusing containers is an experimental feature, so please acknowledge you can experience some issues while using it. If you find any issue, please report it [here](https://github.com/testcontainers/testcontainers-go/issues/new?assignees=&labels=bug&projects=&template=bug_report.yml&title=%5BBug%5D%3A+).

A `ReusableContainer` is a container you mark to be reused across different tests. Reusing containers works out of the box just by setting the `Reuse` field in the `ContainerRequest` to `true`.
Internally, _Testcontainers for Go_ automatically creates a hash of the container request and adds it as a container label. Two labels are added:

- `org.testcontainers.hash` - the hash of the container request.
- `org.testcontainers.copied_files.hash` - the hash of the files copied to the container using the `Files` field in the container request.

!!!info
	Only the files copied in the container request will be checked for reuse. If you copy a file to a container after it has been created, as in the example below, the container will still be reused, because the original request has not changed. Directories added in the `Files` field are not included in the hash, to avoid performance issues calculating the hash of large directories.

If there is no container with those two labels matching the hash values, _Testcontainers for Go_ creates a new container. Otherwise, it reuses the existing one.

This behaviour persists across multiple test runs, as long as the container request remains the same. Ryuk the resource reaper does not terminate that container if it is marked for reuse, as it does not match the prune conditions used by Ryuk. To know more about Ryuk, please read the [Garbage Collector](/features/garbage_collector#ryuk) documentation.

!!!warning
	In the case different test programs are creating a container with the same hash, we must check if the container is already created.
	For that _Testcontainers for Go_ waits up-to 5 seconds for the container to be created. If the container is not found,
	the code proceedes with the creation of the container, else the container is reused.
	This wait is needed because we need to synchronize the creation of the container across different test programs,
	so you could find very rare situations where the container is not found in different test sessions and it is created in them.

### Reuse example

The following example creates an NGINX container, adds a file into it and then reuses the container again for checking the file:

```go
package testcontainers_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleReusableContainer_usingACopiedFile() {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "nginx:1.17.6",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForListeningPort("80/tcp"),
		Reuse:        true, // mark the container as reusable
	}

	n1, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}
	// not terminating the container on purpose, so that it can be reused in a different test.
	// defer n1.Terminate(ctx)

	// Let's copy a file to the container, to demonstrate that successive containers can use the same files
	// when the container is marked for reuse.
	bs := []byte(`#!/usr/bin/env bash
echo "hello world" > /data/hello.txt
echo "done"`)

	copiedFileName := "hello_copy.sh"
	err = n1.CopyToContainer(ctx, bs, "/"+copiedFileName, 700)
	if err != nil {
		log.Fatal(err)
	}

	// Because n2 uses the same container request, it will reuse the container created by n1.
	n2, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}

	c, _, err := n2.Exec(ctx, []string{"bash", copiedFileName})
	if err != nil {
		log.Fatal(err)
	}

	// the file must exist in this second container, as it's reusing the first one
	fmt.Println(c)

	// Output: 0
}

```

Becuase the `Reuse` option is set to `true`, and the copied files have not changed, the container request is the same, resulting in the second container reusing the first one and the file `hello_copy.sh` being executed.

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
