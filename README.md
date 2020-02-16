[![Build Status](https://travis-ci.org/testcontainers/testcontainers-go.svg?branch=master)](https://travis-ci.org/testcontainers/testcontainers-go)

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

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNginxLatestReturn(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
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


## Build from Dockerfile

Testcontainers-go gives you the ability to build an image and run a container from a Dockerfile.

You can do so by specifying a `Context` (the filepath to the build context on your local filesystem) 
and optionally a `Dockerfile` (defaults to "Dockerfile") like so:

```go
req := ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: "/path/to/build/context",
			Dockerfile: "CustomDockerfile",
		},
	}
```

### Dynamic Build Context

If you would like to send a build context that you created in code (maybe you have a dynamic Dockerfile), you can
send the build context as an `io.Reader` since the Docker Daemon accepts is as a tar file, you can use the [tar](https://golang.org/pkg/archive/tar/) package to create your context.


To do this you would use the `ContextArchive` attribute in the `FromDockerfile` struct.

```go
var buf bytes.Buffer
tarWriter := tar.NewWriter(&buf)
// ... add some files
if err := tarWriter.Close(); err != nil {
	// do something with err
}
reader := bytes.NewReader(buf.Bytes())
fromDockerfile := testcontainers.FromDockerfile{
	ContextArchive: reader,
}
```

**Please Note** if you specify a `ContextArchive` this will cause testcontainers to ignore the path passed
in to `Context`

## Sending a CMD to a Container

If you would like to send a CMD (command) to a container, you can pass it in to the container request via the `Cmd` field...

```go
req := ContainerRequest{
	Image: "alpine",
	WaitingFor: wait.ForAll(
		wait.ForLog("command override!"),
	),
	Cmd: []string{"echo", "command override!"},
}
```

## Following Container Logs

If you wish to follow container logs, you can set up `LogConsumer`s.  The log following functionality follows
a producer-consumer model.  You will need to explicitly start and stop the producer.  As logs are written to either
`stdout`, or `stderr` (`stdin` is not supported) they will be forwarded (produced) to any associated `LogConsumer`s.  You can associate `LogConsumer`s
with the `.FollowOutput` function.

**Please note** if you start the producer you should always stop it explicitly.

for example, this consumer will just add logs to a slice

```go
type TestLogConsumer struct {
	Msgs []string
}

func (g *TestLogConsumer) Accept(l Log) {
	g.Msgs = append(g.Msgs, string(l.Content))
}
```

this can be used like so:
```go
g := TestLogConsumer{
	Msgs: []string{},
}

err := c.StartLogProducer(ctx)
if err != nil {
	// do something with err
}

c.FollowOutput(&g)

// some stuff happens...

err = c.StopLogProducer()
if err != nil {
	// do something with err
}
```
