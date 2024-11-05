# Health Wait strategy

The health wait strategy will check that the container is in the healthy state and allows to set the following conditions:

- the startup timeout to be used in seconds, default is 60 seconds.
- the poll interval to be used in milliseconds, default is 100 milliseconds.

```golang
req := ContainerRequest{
	Image:      "alpine:latest",
	WaitingFor: wait.ForHealthCheck(),
}
```
