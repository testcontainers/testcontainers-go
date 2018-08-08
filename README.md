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
	testcontainer "github.com/gianarb/testcontainer-go"
	"net/http"
	"testing"
)

func TestNginxLatestReturn(t *testing.T) {
	ctx := context.Background()
	nginxC, err := testcontainer.RunContainer(ctx, "nginx", testcontainer.RequestContainer{
		ExportedPort: []string{
			"80/tcp",
		},
	})
	if err != nil {
		t.Error(err)
	}
	defer nginxC.Terminate(ctx, t)
	ip, err := nginxC.GetIPAddress(ctx)
	if err != nil {
		t.Error(err)
	}
	resp, err := http.Get(fmt.Sprintf("http://%s", ip))
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}
```
This is a simple example, you can create one container in my case using the
`nginx` image. You can get its IP `ip, err := nginxC.GetIPAddress(ctx)` and you
can use it to make a GET: `resp, err := http.Get(fmt.Sprintf("http://%s", ip))`

To clean your environment you can defer the container termination `defer
nginxC.Terminate(ctx, t)`. `t` is `*testing.T` and it is used to notify is the
`defer` failed marking the test as failed.

You can build more complex flow using envvar to configure the containers. Let's
suppose you are testing an application that requites redis:

```go
func TestRedisPing(t testing.T) {
    ctx := context.Background()
    redisC, err := testcontainer.RunContainer(ctx, "redis", testcontainer.RequestContainer{
        ExportedPort: []string{
            "6379/tcp",
        },
    })
    if err != nil {
        t.Error(err)
    }
    defer redisC.Terminate(ctx, t)
    redisIP, err := redisC.GetIPAddress(ctx)
    if err != nil {
        t.Error(err)
    }

    appC, err := testcontainer.RunContainer(ctx, "your/app", testcontainer.RequestContainer{
        ExportedPort: []string{
            "8081/tcp",
        },
        Env: map[string]string{
            "REDIS_HOST": fmt.Sprintf("http://%s:6379", redisIP),
        },
    })
    if err != nil {
        t.Error(err)
    }
    defer appC.Terminate(ctx, t)

    // your assertions
}
```
