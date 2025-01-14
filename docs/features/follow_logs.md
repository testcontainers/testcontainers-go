# Following Container Logs

The log-following functionality follows a producer-consumer model: the container produces logs, and your code consumes them.
So if you wish to follow container logs, you have to do two things:

1. set up log consumers.
2. configure the log production of the container (e.g. timeout for the logs).

As logs are written to either `stdout`, or `stderr` (`stdin` is not supported) they will be forwarded (produced) to any associated log consumer.

## Creating a LogConsumer

A `LogConsumer` must implement the `LogConsumer` interface, and it could be as simple as directly printing the log to `stdout`,
as in the following example:

<!--codeinclude-->
[The LogConsumer Interface](../../logconsumer.go) inside_block:logConsumerInterface
[The Log struct](../../logconsumer.go) inside_block:logStruct
[Example LogConsumer](../../testing.go) inside_block:exampleLogConsumer
<!--/codeinclude-->

You can associate `LogConsumer`s in two manners:

1. as part of the `ContainerRequest` struct.
2. with the `FollowOutput` function (deprecated).

## Passing the LogConsumers in the ContainerRequest

This will represent the current way for associating `LogConsumer`s. You simply define your consumers, and attach them as a slice to the `ContainerRequest` in the
`LogConsumerCfg` field. See the following example, where `g` is an instance of a given `LogConsumer` struct.

<!--codeinclude-->
[Passing LogConsumers](../../logconsumer_test.go) inside_block:logConsumersAtRequest
<!--/codeinclude-->

Please check that it's possible to configure the log production with a slice of functional options. These options must be of the `LogProductionOption` type:

```go
type LogProductionOption func(*DockerContainer)
```

At the moment, _Testcontainers for Go_ exposes an option to set log production timeout, using the `WithLogProductionTimeout` function.

_Testcontainers for Go_ will read this log producer/consumer configuration to automatically start producing logs if and only if the consumers slice contains at least one valid `LogConsumer`.

## Manually using the FollowOutput function

!!!warning
	This method is not recommended, as it requires you to manually manage the `LogConsumer` lifecycle.
	We recommend using the `ContainerRequest` struct to associate `LogConsumer`s, as it's the simplest and most straightforward method.
	If you use both methods, you can get an error, as the `StartLogProducer` function could be called twice, which is not allowed.

	As a consequence, this lifecycle (`StartLogProducer`, `FollowOutput` and `StopLogProducer) will be **deprecated** in the future, delegating the control to the library.

Instead of passing the `LogConsumer` as part of the `ContainerRequest` struct, you can manually call the `FollowOutput` function on a `Container` instance.
This allows you to dynamically add `LogConsumer`s to a running container, although it forces you to manually manage the `LogConsumer` lifecycle,
calling `StartLogProducer` **after** the `FollowOutput` function, and do it just once.

You can define a log consumer like so:

```go
type TestLogConsumer struct {
	Msgs []string // store the logs as a slice of strings
}

func (g *TestLogConsumer) Accept(l testcontainers.Log) {
	g.Msgs = append(g.Msgs, string(l.Content))
}
```

And then associate it with a container like so:

```go
g := TestLogConsumer{
	Msgs: []string{},
}

// Remember that this method will be deprecated in the future
c.FollowOutput(&g) // must be called before StarLogProducer

// Remember that this method will be deprecated in the future
err := c.StartLogProducer(ctx)
if err != nil {
	// do something with err
}

// some stuff happens...

// Remember that this method will be deprecated in the future
err = c.StopLogProducer()
if err != nil {
	// do something with err
}
```

## Stopping the Log Production

The production of logs is automatically stopped in `c.Terminate()`, so you don't have to worry about that.

!!! warning
	It can be done manually during container lifecycle using `c.StopLogProducer()`, but it's not recommended, as it will be deprecated in the future.

## Listening to errors

When the log production fails to start within given timeout (causing a context deadline) or there's an error returned while closing the reader it will no longer panic, but instead will return an error over a channel. You can listen to it using `DockerContainer.GetLogProductionErrorChannel()` method:

```go
func (c *DockerContainer) GetLogProductionErrorChannel() <-chan error {
	return c.producerError
}
```

This allows you to, for example, retry restarting the log production if it fails to start the first time.

For example, you would start the log production normally, defining the log production configuration at the `ContainerRequest` struct, and then:

```go
// start log production normally, using the ContainerRequest struct, or
// using the deprecated c.StartLogProducer method.
// err = container.StartLogProducer(ctx, WithLogProductionTimeout(10*time.Second))

// listen to errors in a detached goroutine
go func(done chan struct{}, timeout time.Duration) {
	for {
		select {
		case logErr := <-container.GetLogProductionErrorChannel():
			if logErr != nil {
				// do something with error
				// for example, retry starting the log production 
				// (here we retry it once, in real life you might want to retry it more times)
				startErr := container.StartLogProducer(ctx, timeout)
				if startErr != nil {
					return 
				}
		case <-done:
			return
		}
	}
}(cons.logListeningDone, time.Duration(10*time.Second))
```
