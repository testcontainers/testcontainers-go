# HostPort Wait strategy

The host-port wait strategy will check if the container is listening to a specific port and allows to set the following conditions:

- a port exposed by the container. The port and protocol to be used, which is represented by a string containing the port number and protocol in the format "80/tcp".
- the first exposed port in the container.
- the startup timeout to be used, default is 60 seconds.
- the poll interval to be used, default is 100 milliseconds.

Variations on the HostPort wait strategy are supported, including:

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
