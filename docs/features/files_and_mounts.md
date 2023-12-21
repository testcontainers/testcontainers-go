# Copying data into a container

Copying data of any type into a container is a very common practice when working with containers. This section will show you how to do it using _Testcontainers for Go_.

## Volume mapping

It is possible to map a Docker volume into the container using the `Mounts` attribute at the `ContainerRequest` struct. For that, please pass an instance of the `GenericVolumeMountSource` type, which allows you to specify the name of the volume to be mapped, and the path inside the container where it should be mounted:

<!--codeinclude-->
[Volume mounts](../../mounts_test.go) inside_block:volumeMounts
<!--/codeinclude-->

!!!tip
    This ability of creating volumes is also available for remote Docker hosts.

!!!warning
    Bind mounts are not supported, as it could not work with remote Docker hosts.

## Copying files to a container

If you would like to copy a file to a container, you can do it in two different manners:

1. Adding a list of files in the `ContainerRequest`, which will be copied before the container starts:

```go
ctx := context.Background()

nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
			Files: []ContainerFile{
				{
					HostFilePath:      "./testdata/hello.sh",
					ContainerFilePath: "/copies-hello.sh",
					FileMode:          0o700,
				},
			},
		},
		Started: false,
	})
```

2. Using the `CopyFileToContainer` method on a `running` container:

```go
ctx := context.Background()

nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
		},
		Started: true,
	})

nginxC.CopyFileToContainer(ctx, "./testdata/hello.sh", "/hello_copy.sh", 0o700)
```

## Copying directories to a container

It's also possible to copy an entire directory to a container, and that can happen before and/or after the container gets into the `Running` state. As an example, you could need to bulk-copy a set of files, such as a configuration directory that does not exist in the underlying Docker image.

It's important to notice that, when copying the directory to the container, the container path must exist in the Docker image. And this is a strong requirement for files to be copied _before_ the container is started, as we cannot create the full path at that time.

You can leverage the very same mechanism used for copying files to a container, but for directories.:

1. The first way is using the `Files` field in the `ContainerRequest` struct, as shown in the previous section, but using the path of a directory as `HostFilePath`.

2. The second way uses the existing `CopyFileToContainer` method, which will internally check if the host path is a directory, calling the `CopyDirToContainer` method if needed:

```go
ctx := context.Background()
// as the container is started, we can create the directory first
_, _, err = myContainer.Exec(ctx, []string{"mkdir", "-p", "/usr/lib/my-software/config"})
// because the container path is a directory, it will use the copy dir method as fallback
err = myContainer.CopyFileToContainer(ctx, "./files", "/usr/lib/my-software/config/files", 0o700)
if err != nil {
	// handle error
}
```

3. The last third way uses the `CopyDirToContainer` method, directly, which, as you probably know, needs the existence of the parent directory in order to copy the directory:

```go
ctx := context.Background()

// as the container is started, we can create the directory first
_, _, err = nginxC.Exec(ctx, []string{"mkdir", "-p", "/usr/lib/my-software/config"})
err = nginxC.CopyDirToContainer(ctx, "./plugins", "/usr/lib/my-software/config/plugins", 0o700)
if err != nil {
	// handle error
}
```
