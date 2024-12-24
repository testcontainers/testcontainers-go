# Test Session Semantics

The test session semantics is a feature that allows _Testcontainers for Go_ to identify the current test session
and tag the containers created by the library with a unique session ID.

This is needed because each Go package will be run in a separate process, so we need a way to identify the current test execution
to aggregate the tests executed in it.

By test session, we mean:

- a single `go test` invocation (including flags).
- a single `go test ./...` invocation, for all subpackages from that location (including flags).
- the execution of a single test or a set of tests using the IDE.

As a consequence, _Testcontainers for Go_ will use the parent process ID (pid) of the current process and its creation date
to generate a unique session ID.

We are using the parent pid because the current `go test` process running a given Go package will be a child of one of the following:

- the process that is running the tests, e.g.: `go test`;
- the process that is running the application in development mode, e.g. `go run main.go -tags dev`;
- the process that is running the tests in the IDE, e.g.: `go test ./...`.

That's why we need to use the parent pid to identify the current test session, as it must be unique.

Finally, we will hash the combination of the `testcontainers-go:` string with the parent pid and the creation date
of that parent process to generate a unique session ID.

After that, the `sessionID` will be used to:

- identify the test session, aggregating the test execution of multiple packages in the same test session.
- pass the `sessionID` to the container runtime, as an HTTP header to the daemon.
- tag the containers created by _Testcontainers for Go_, adding a label to the container with this session ID.
