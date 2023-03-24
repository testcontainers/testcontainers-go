# Copy Files To Container

If you would like to copy a file to a container, you can do it using the `CopyFileToContainer` method...

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

nginxC.CopyFileToContainer(ctx, "./testdata/hello.sh", "/hello_copy.sh", 700)
```

Or you can add a list of files in the `ContainerRequest` initialization, which can be copied before the container starts:

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
					FileMode:          700,
				},
			},
		},
		Started: false,
	})
```

## Copy Directories To Container

It's also possible to copy an entire directory to a container, and that can happen before and/or after the container gets into the "Running" state. As an example, you could need to bulk-copy a set of files, such as a configuration directory that does not exist in the underlying Docker image.

It's important to notice that, when copying the directory to the container, the container path must exist in the Docker image. And this is a strong requirement for files to be copied _before_ the container is started, as we cannot create the full path at that time.

There are two ways to copy directories to a container. The first way uses the existing `CopyFileToContainer` method, which will internally check if the host path is a directory, internally calling the new `CopyDirToContainer` method if needed:

```go
ctx := context.Background()

// copy a directory before the container is started, using Files field
nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
			Files: []ContainerFile{
				{
					HostFilePath:      "./testdata",    // a directory
					ContainerFilePath: "/tmp/testdata", // important! its parent already exists
					FileMode:          700,
				},
			},
		},
		Started: true,
	})
if err != nil {
	// handle error
}

// as the container is started, we can create the directory first
_, _, err = nginxC.Exec(ctx, []string{"mkdir", "-p", "/usr/lib/my-software/config"})
// because the container path is a directory, it will use the copy dir method as fallback
err = nginxC.CopyFileToContainer(ctx, "./files", "/usr/lib/my-software/config/files", 700)
if err != nil {
	// handle error
}
```

And the second way uses the `CopyDirToContainer` method which, as you probably know, needs the existence of the parent directory in order to copy the directory:

```go
ctx := context.Background()

// copy a directory before the container is started, using Files field
nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
			Files: []ContainerFile{
				{
					HostFilePath:      "./testdata",    // a directory
					ContainerFilePath: "/tmp/testdata", // important! its parent already exists
					FileMode:          700,
				},
			},
		},
		Started: true,
	})
if err != nil {
	// handle error
}

// as the container is started, we can create the directory first
_, _, err = nginxC.Exec(ctx, []string{"mkdir", "-p", "/usr/lib/my-software/config"})
err = nginxC.CopyDirToContainer(ctx, "./plugins", "/usr/lib/my-software/config/plugins", 700)
if err != nil {
	// handle error
}
```
