# Exit Wait strategy

The exit wait strategy will check the container is not in the running state, being able to set the following conditions:

- the exit timeout, default is `0`.
- the PollInterval to be used, default is 100 milliseconds

## Match an exit code

```golang
req := ContainerRequest{
	Image:      "docker.io/alpine:latest",
	WaitingFor: wait.ForExit(),
}
```
