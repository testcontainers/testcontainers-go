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

