# Exit Wait strategy

The exit wait strategy will check that the container is not in the running state, and allows to set the following conditions:

- the exit timeout in seconds, default is `0`.
- the poll interval to be used in milliseconds, default is 100 milliseconds.

## Match an exit code

```golang
req := ContainerRequest{
	Image:      "alpine:latest",
	WaitingFor: wait.ForExit(),
}
```
