# Executing commands

## Container startup command

By default the container will execute whatever command is specified in the image's Dockerfile. If you would like to send a CMD (command) to a container, you can pass it in to
the container request via the `Cmd` field. For example:

```go
req := ContainerRequest{
	Image: "alpine",
	WaitingFor: wait.ForAll(
		wait.ForLog("command override!"),
	),
	Cmd: []string{"echo", "command override!"},
}
```

!!!info
    If you are using a module, you can use the `testcontainers.CustomizeRequest` option to add arguments to the command. Check the individual module's pages for more information on their commands.

This option will merge the customized request into the module's request, appending any additional `Cmd` arguments to the
module's command. This can't be used to replace the command, only to append options.
Check the individual module's pages for more information on their commands.

## Executing a command

You can execute a command inside a running container, similar to a `docker exec` call:

<!--codeinclude-->
[Executing a command](../../docker_test.go) inside_block:exec_example
<!--/codeinclude-->

This can be useful for software that has a command line administration tool. You can also get the logs of the command execution (from an object that implements the [io.Reader](https://pkg.go.dev/io#Reader) interface). For example:


<!--codeinclude-->
[Command logs](../../docker_test.go) inside_block:exec_reader_example
<!--/codeinclude-->

This is done this way, because it brings more flexibility to the user, rather than returning a string.
