When I was working on a Zipkin PR I discovered a nice Java library called
[testcontainers](https://www.testcontainers.org/).

It provides an easy and clean API over the go docker sdk to run, terminate and
connect to containers in your tests.

I found myself comfortable programmatically writing the containers I need to run
an integration/smoke tests. So I started porting this library in Go.


This is the API I have defined:

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	testcontainer "github.com/testcontainers/testcontainer-go"
)

func TestNginxLatestReturn(t *testing.T) {
	ctx := context.Background()
	req := testcontainer.ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
	}
	nginxC, err := testcontainer.GenericContainer(ctx, testcontainer.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	defer nginxC.Terminate(ctx)
	ip, err := nginxC.Host(ctx)
	if err != nil {
		t.Error(err)
	}
	port, err := nginxC.MappedPort(ctx, "80")
	if err != nil {
		t.Error(err)
	}
	resp, err := http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}
```
This is a simple example, you can create one container in my case using the
`nginx` image. You can get its IP `ip, err := nginxC.GetContainerIpAddress(ctx)` and you
can use it to make a GET: `resp, err := http.Get(fmt.Sprintf("http://%s", ip))`

To clean your environment you can defer the container termination `defer
nginxC.Terminate(ctx, t)`. `t` is `*testing.T` and it is used to notify is the
`defer` failed marking the test as failed.

You can build more complex flow using envvar to configure the containers. Let's
suppose you are testing an application that requites redis:

```go
func TestRedisPing(t *testing.T) {
	ctx := context.Background()
	req := testcontainer.ContainerRequest{
		Image:        "redis",
		ExposedPorts: []string{"6379/tcp"},
	}
	redisC, err := testcontainer.GenericContainer(ctx, testcontainer.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	defer redisC.Terminate(ctx)
	ip, err := redisC.Host(ctx)
	if err != nil {
		t.Error(err)
	}
	redisPort, err := redisC.MappedPort(ctx, "6479/tcp")
	if err != nil {
		t.Error(err)
	}

	appReq := testcontainer.ContainerRequest{
		ExposedPorts: []string{"8081/tcp"},
		Env: map[string]string{
			"REDIS_HOST": fmt.Sprintf("http://%s:%s", ip, redisPort.Port()),
		},
	}
	appC, err := testcontainer.RunContainer(ctx, "your/app", testcontainer.GenericContainerRequest{
		ContainerRequest: appReq,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	defer appC.Terminate(ctx)

	// your assertions
}
```
