# Build from Dockerfile

_Testcontainers for Go_ gives you the ability to build an image and run a container
from a Dockerfile.

You can do so by specifying a `Context` (the filepath to the build context on
your local filesystem) and optionally a `Dockerfile` (defaults to "Dockerfile")
like so:

<!--codeinclude-->
[Building From a Dockerfile including Repository and Tag](../../from_dockerfile_test.go) inside_block:fromDockerfileIncludingRepo
<!--/codeinclude-->

As you can see, you can also specify the `Repo` and `Tag` optional fields to use for the image. If not passed, the
image will be built with a random name and tag.

If your Dockerfile expects build args: 

```Dockerfile
FROM alpine

ARG FOO

```
You can specify them like:

<!--codeinclude-->
[Building From a Dockerfile including build arguments](../../docker_test.go) inside_block:fromDockerfileWithBuildArgs
<!--/codeinclude-->

## Dynamic Build Context

If you would like to send a build context that you created in code (maybe you have a dynamic Dockerfile), you can
send the build context as an `io.Reader` since the Docker Daemon accepts it as a tar file, you can use the [tar](https://golang.org/pkg/archive/tar/) package to create your context.


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

**Please Note** if you specify a `ContextArchive` this will cause _Testcontainers for Go_ to ignore the path passed
in to `Context`.

## Images requiring auth

If you are building a local Docker image that is fetched from a Docker image in a registry requiring authentication
(e.g., assuming you are fetching from a custom registry such as `myregistry.com`), _Testcontainers for Go_ will automatically
discover the credentials for the given Docker image from the Docker config, as described [here](./docker_auth.md).

```go
req := ContainerRequest{
    FromDockerfile: testcontainers.FromDockerfile{
        Context: "/path/to/build/context",
        Dockerfile: "CustomDockerfile",
	},
}
```

## Keeping built images

Per default, built images are deleted after being used.
However, some images you build might have no or only minor changes during development.
Building them for each test run might take a lot of time.
You can avoid this by setting `KeepImage` in `FromDockerfile`.
If the image is being kept, cached layers might be reused during building or even the whole image.

```go
req := ContainerRequest{
    FromDockerfile: testcontainers.FromDockerfile{
        // ...
		KeepImage: true,
	},
}
```

## Advanced usage

In the case you need to pass additional arguments to the `docker build` command, you can use the `BuildOptionsModifier` attribute in the `FromDockerfile` struct.

This field holds a function that has access to Docker's ImageBuildOptions type, which is used to build the image. You can use this modifier **on your own risk** to modify the build options with as many options as you need.

<!--codeinclude-->
[Building From a Dockerfile including build options modifier](../../from_dockerfile_test.go) inside_block:buildFromDockerfileWithModifier
[Dockerfile including target](../../testdata/target.Dockerfile)
<!--/codeinclude-->
