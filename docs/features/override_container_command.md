# Executing commands

## Container startup command

By default the container will execute whatever command is specified in the image's Dockerfile. If you would like to send a CMD (command) to a container, you can pass it in to
the container request via the `Cmd` field. For example:

```go
req := ContainerRequest{
	Image: "alpine",
	WaitingFor: wait.ForAll(
		wait.ForLog("command override!"),
	),
	Cmd: []string{"echo", "command override!"},
}
```

## Executing a command

You can execute a command inside a running container, similar to a `docker exec` call:

```go
func TestIntegrationNginxLatestReturn(t *testing.T) {
    ctx := context.Background()
    req := ContainerRequest{
    Image: "docker.io/busybox",
        Cmd:   []string{"sleep", "10"},
           Tmpfs: map[string]string{"/testtmpfs": "rw"},
    }

    container, err := GenericContainer(ctx, GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })

    require.NoError(t, err)

    t.Cleanup(func() {
        t.Log("terminating container")
        require.NoError(t, container.Terminate(ctx))
    })

    path := "/testtmpfs/test.file"

    c, _, err := container.Exec(ctx, []string{"ls", path})
    if err != nil {
        t.Fatal(err)
    }
}
```

This can be useful for software that has a command line administration tool. You can also get the logs of the command execution (from an object that implements the [io.Reader](https://pkg.go.dev/io#Reader) interface). For example:

```go
import (
    "fmt"
    "io"
    "log"
    "strings"
    "testing"
)

func TestIntegrationNginxLatestReturn(t *testing.T) {
    // ...

    c, reader, err := container.Exec(ctx, []string{"ls", path})
    if err != nil {
        t.Fatal(err)
    }

    buf := new(strings.Builder)
    _, err := io.Copy(buf, reader)
    if err != nil {
        t.Fatal(err)
    }

    // See the logs of the command execution.
    t.Log(buf.String())
}
```

This is done this way, because it brings more flexibility to the user, rather than returning a string.
