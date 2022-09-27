# Health Wait strategy

The health wait strategy will check the container is in the healthy state, being able to set the following conditions:

- the startupTimeout to be used, default is 60 seconds.
- the PollInterval to be used, default is 100 milliseconds.

```golang
req := ContainerRequest{
	Image:      "docker.io/alpine:latest",
	WaitingFor: wait.ForHealthCheck(),
}
```
