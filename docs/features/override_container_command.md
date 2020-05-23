# Sending a CMD to a Container

If you would like to send a CMD (command) to a container, you can pass it in to
the container request via the `Cmd` field...

```go
req := ContainerRequest{
	Image: "alpine",
	WaitingFor: wait.ForAll(
		wait.ForLog("command override!"),
	),
	Cmd: []string{"echo", "command override!"},
}
```

