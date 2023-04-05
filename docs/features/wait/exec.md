# Exec Wait Strategy

The exec wait strategy will check the exit code of a process to be executed in the container, and allows to set the following conditions:

- the command and arguments to be executed, as an array of strings.
- a function to match a specific exit code, with the default matching `0`.
- the output response matcher as a function.
- the startup timeout to be used in seconds, default is 60 seconds.
- the poll interval to be used in milliseconds, default is 100 milliseconds.

## Match an exit code and a response matcher

<!--codeinclude-->
[Waiting for a command matching an exit code and response](../../../wait/exec_test.go) inside_block:waitForExecExitCodeResponse
<!--/codeinclude-->
