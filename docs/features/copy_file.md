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

nginxC.CopyFileToContainer(ctx, "./testresources/hello.sh", "/hello_copy.sh", fileContent, 700)
```

Or you can add a list of files in ContainerRequest's struct, which can be copied before the container started:

```go
ctx := context.Background()

nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
			Files: []ContainerFile{
				{
					HostFilePath:      "./testresources/hello.sh",
					ContainerFilePath: "/copies-hello.sh",
					FileMode:          700,
				},
			},
		},
		Started: false,
	})
```

