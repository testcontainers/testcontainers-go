# Wait Strategies

There are scenarios where your tests need the external services they rely on to reach a specific state that is particularly useful for testing. This is generally approximated as 'Can we talk to this container over the network?' or 'Let's wait until the container is running and reaches certain state'.

_Testcontainers for Go_ comes with the concept of `wait strategy`, which allows your tests to actually wait for the most useful conditions to be met, before continuing with their execution. These wait strategies are implemented in the `wait` package.

Below you can find a list of the available wait strategies that you can use:

- [Exec](./exec.md)
- [Exit](./exit.md)
- [File](./file.md)
- [Health](./health.md)
- [HostPort](./host_port.md)
- [HTTP](./http.md)
- [Log](./log.md)
- [Multi](./multi.md)
- [SQL](./sql.md)
- [TLS](./tls.md)

## Startup timeout and Poll interval

When defining a wait strategy, it should define a way to set the startup timeout to avoid waiting infinitely. For that, _Testcontainers for Go_ creates a cancel context with 60 seconds defined as timeout.

If the default 60s timeout is not sufficient, it can be updated with the `WithStartupTimeout(startupTimeout time.Duration)` function.

Besides that, it's possible to define a poll interval, which will actually stop 100 milliseconds the test execution.

If the default 100 milliseconds poll interval is not sufficient, it can be updated with the `WithPollInterval(pollInterval time.Duration)` function.

## Modifying request strategies

It's possible for options to modify `ContainerRequest.WaitingFor` using
[Walk](walk.md).
