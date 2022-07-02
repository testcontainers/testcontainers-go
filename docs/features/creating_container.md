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
	defer nginxC.Terminate(ctx)

	resp, err := http.Get(nginxC.URI)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}
```

## Reusable container

With `Reuse` option you can reuse an existing container. Reusing will work only if you pass an 
existing container name via 'req.Name' field. If the name is not in a list of existing containers, 
the function will create a new generic container. If `Reuse` is true and `Name` is empty, you will get error.

The following test creates an NGINX container, adds a file into it and then reuses the container again for checking the file:
```go

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
err = n1.CopyFileToContainer(ctx, "./testresources/hello.sh", "/"+copiedFileName, 700)

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