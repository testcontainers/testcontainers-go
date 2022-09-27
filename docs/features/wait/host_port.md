# HostPort wait strategy

You can choose to wait for:

- a port to be listening in the container. The port and protocol to be used, which is represented by a string containing port number and protocol in the format "80/tcp"
- the first exposed port in the container.

## Waiting for a port in the container

```golang
req := ContainerRequest{
    Image:        "docker.io/nginx:alpine",
    ExposedPorts: []string{"80/tcp"},
    WaitingFor:   wait.ForListeningPort("80/tcp"),
}
```

## Waiting for the first exposed port in the container

The wait strategy will use the first exposed port from the image configuration.

```golang
req := ContainerRequest{
    Image:        "docker.io/nginx:alpine",
    WaitingFor:   wait.ForExposedPort(),
}
```
