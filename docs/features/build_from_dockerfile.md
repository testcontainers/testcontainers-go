# Build from Dockerfile

Testcontainers-go gives you the ability to build an image and run a container
from a Dockerfile.

You can do so by specifying a `Context` (the filepath to the build context on
your local filesystem) and optionally a `Dockerfile` (defaults to "Dockerfile")
like so:

```go
req := ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: "/path/to/build/context",
			Dockerfile: "CustomDockerfile",
		},
	}
```

If your Dockerfile expects build args: 

```Dockerfile
FROM alpine

ARG FOO

```
You can specify them like:

```go
req := ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: "/path/to/build/context",
			Dockerfile: "CustomDockerfile",
			BuildArgs: map[string]*string {
				"FOO": "BAR",
			},
		},
	}
```
## Dynamic Build Context

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
