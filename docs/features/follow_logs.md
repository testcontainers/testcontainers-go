# Following Container Logs

The log-following functionality follows a producer-consumer model: the container produces logs, and your code consumes them.
So if you wish to follow container logs, you have to do two things:

1. set up log consumers.
2. configure the log production of the container (e.g. timeout for the logs).

As logs are written to either `stdout`, or `stderr` (`stdin` is not supported) they will be forwarded (produced) to any associated log consumer.

## Creating a LogConsumer

A `LogConsumer` must implement the `log.Consumer` interface, and it could be as simple as directly printing the log to `stdout`,
as in the following example:

<!--codeinclude-->
[The LogConsumer Interface](../../log/consumer.go) inside_block:logConsumerInterface
[The Log struct](../../log/log.go) inside_block:logStruct
[Example LogConsumer](../../log/consumer.go) inside_block:exampleLogConsumer
<!--/codeinclude-->

You can associate `LogConsumer`s as part of the `Request` struct.

## Passing the LogConsumers in the Request

This will represent the current way for associating `LogConsumer`s. You simply define your consumers, and attach them as a slice to the `Request` in the
`LogConsumerCfg` field. See the following example, where `g` is an instance of a given `LogConsumer` struct.

<!--codeinclude-->
[Passing LogConsumers](../../logconsumer_test.go) inside_block:logConsumersAtRequest
<!--/codeinclude-->

In order to provide multiple log consumers, you can use the `log.MultiConsumer` composite struct, which will forward logs to all of its consumers.

<!--codeinclude-->
[Composite Log Consumer](../../log/consumer.go) inside_block:multiLogConsumer
<!--/codeinclude-->

Please check that it's possible to configure the log production with an slice of functional options. These options must be of the `log.ProductionOption` type:

```go
type ProductionOption func(*DockerContainer)
```

At the moment, _Testcontainers for Go_ exposes an option to set log production timeout, using the `WithProductionTimeout` function.

_Testcontainers for Go_ will read this log producer/consumer configuration to automatically start producing logs if an only if the consumers slice contains at least one valid `LogConsumer`.

## Stopping the Log Production

The production of logs is automatically stopped in `c.Terminate()`, so you don't have to worry about that.

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
// err = container.StartLogProducer(ctx, WithProductionTimeout(10*time.Second))

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
