# HostPort Wait strategy

The host-port wait strategy will check if the container is listening to a specific port and allows to set the following conditions:

- a port exposed by the container. The port and protocol to be used, which is represented by a string containing the port number and protocol in the format "80/tcp".
- alternatively, wait for the lowest exposed port in the container.
- the startup timeout to be used, default is 60 seconds.
- the poll interval to be used, default is 100 milliseconds.
- skip the internal check.

Variations on the HostPort wait strategy are supported, including:

## Listening port in the container

```golang
req := ContainerRequest{
    Image:        "nginx:alpine",
    ExposedPorts: []string{"80/tcp"},
    WaitingFor:   wait.ForListeningPort("80/tcp"),
}
```

## Lowest exposed port in the container

The wait strategy will use the lowest exposed port from the container configuration.

```golang
req := ContainerRequest{
    Image:        "nginx:alpine",
    WaitingFor:   wait.ForExposedPort(),
}
```

Said that, it could be the case that the container request included ports to be exposed. Therefore using `wait.ForExposedPort` will wait for the lowest exposed port in the request, because the container configuration retrieved from Docker will already include them.

```golang
req := ContainerRequest{
    Image:        "nginx:alpine",
    ExposedPorts: []string{"80/tcp", "9080/tcp"},
    WaitingFor:   wait.ForExposedPort(),
}
```

## Skipping the internal check

_Testcontainers for Go_ checks if the container is listening to the port internally before returning the control to the caller. For that it uses a shell command to check the port status:

<!--codeinclude-->
[Internal check](../../../wait/host_port.go) inside_block:buildInternalCheckCommand
<!--/codeinclude-->

But there are cases where this internal check is not needed, for example when a shell is not available in the container or
when the container doesn't bind the port internally until additional conditions are met.
In this case, the `wait.ForExposedPort.SkipInternalCheck` can be used to skip the internal check.

```golang
req := ContainerRequest{
    Image:        "nginx:alpine",
    ExposedPorts: []string{"80/tcp", "9080/tcp"},
    WaitingFor:   wait.ForExposedPort().SkipInternalCheck(),
}
```

## Skipping the external check

_Testcontainers for Go_ checks if the container is listening to the port externally (outside of container, 
from the host where _Testcontainers for Go_ is used) before returning the control to the caller.

But there are cases where this external check is not needed.
In this case, the `wait.ForListeningPort.SkipExternalCheck` can be used to skip the external check.

```golang
req := ContainerRequest{
    Image:      "nginx:alpine",
    // Do not check port 80 externally, check it internally only
    WaitingFor: wait.ForListeningPort("80/tcp").SkipExternalCheck(),
}
```

If there is a need to wait only for completion of container port mapping (which doesn't happen immediately after container is started),
then both internal and external checks can be skipped:

```golang
req := ContainerRequest{
    Image:        "nginx:alpine",
    ExposedPorts: []string{"80/tcp"},
    // Wait only for completion of port 80 mapping (from container runtime perspective), do not connect to 80 port
    WaitingFor:   wait.ForListeningPort("80/tcp").SkipInternalCheck().SkipExternalCheck(),
}
```

Alternatively, `wait.ForMappedPort` can be used:

```golang
req := ContainerRequest{
    Image:        "nginx:alpine",
    ExposedPorts: []string{"80/tcp"},
    // Wait only for completion of port 80 mapping (from container runtime perspective), do not connect to 80 port
    WaitingFor:   wait.ForMappedPort("80/tcp"),
}
```
