# Exec Wait strategy

The exec wait strategy will check the exit code of a process to be executed in the container, being able to set the following conditions:

- the command and arguments to be executed, as an array of strings
- the exit code as a function to resolve a matcher, being the default one `0`.
- the PollInterval to be used, default is 100 milliseconds

## Match an exit code

```golang
req := ContainerRequest{
	Image:        "docker.io/nginx:alpine",
	WaitingFor: wait.NewExecStrategy([]string{"git", "version"}).WithExitCodeMatcher(func(exitCode int) bool {
		return exitCode == 10
	}),
}
```
