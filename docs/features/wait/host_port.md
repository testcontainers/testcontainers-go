# HostPort wait strategy

The host-port wait strategy will check if a port is listening in the container, being able to set the following conditions:

- a port to be listening in the container. The port and protocol to be used, which is represented by a string containing port number and protocol in the format "80/tcp"
- the first exposed port in the container.
- the startup timeout to be used, default is 60 seconds
- the poll interval to be used, default is 100 milliseconds

Variations on the HosPort wait strategy are supported, including:

## Listening port in the container

```golang
req := ContainerRequest{
    Image:        "docker.io/nginx:alpine",
    ExposedPorts: []string{"80/tcp"},
    WaitingFor:   wait.ForListeningPort("80/tcp"),
}
```

## First exposed port in the container

The wait strategy will use the first exposed port from the image configuration.

```golang
req := ContainerRequest{
    Image:        "docker.io/nginx:alpine",
    WaitingFor:   wait.ForExposedPort(),
}
```
