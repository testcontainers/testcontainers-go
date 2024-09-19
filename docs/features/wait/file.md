# File Wait Strategy

File Wait Strategy waits for a file to exist in the container, and allows to set the following conditions:

- the file to wait for.
- a matcher which reads the file content, no-op if nil or not set.
- the startup timeout to be used in seconds, default is 60 seconds.
- the poll interval to be used in milliseconds, default is 100 milliseconds.

## Waiting for file to exist and extract the content

<!--codeinclude-->
[Waiting for file to exist and extract the content](../../../wait/file_test.go) inside_block:waitForFileWithMatcher
<!--/codeinclude-->
