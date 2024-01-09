# Following Container Logs

If you wish to follow container logs, you can set up `LogConsumer`s.  The log
following functionality follows a producer-consumer model. As logs are written to either `stdout`,
or `stderr` (`stdin` is not supported) they will be forwarded (produced) to any
associated `LogConsumer`s.  You can associate `LogConsumer`s with the
`.FollowOutput` function.

For example, this consumer will just add logs to a slice

```go
type TestLogConsumer struct {
	Msgs []string
}

func (g *TestLogConsumer) Accept(l Log) {
	g.Msgs = append(g.Msgs, string(l.Content))
}
```
This can be used like so:
```go
g := TestLogConsumer{
	Msgs: []string{},
}

c.FollowOutput(&g) // must be called before StarLogProducer

err := c.StartLogProducer(ctx)
if err != nil {
	// do something with err
}

// some stuff happens...

err = c.StopLogProducer()
if err != nil {
	// do something with err
}
```

`LogProducer` is stopped in `c.Terminate()`. It can be done manually during container lifecycle
using `c.StopLogProducer()`. For a particular container, only one `LogProducer` can be active at time.

`StartLogProducer()` also accepts a functional parameter now used to set log producer timeout:
```golang
type LogProducerOption func(*DockerContainer)

func WithLogProducerTimeout(timeout time.Duration) LogProducerOption {
	return func(c *DockerContainer) {
		c.producerTimeout = &timeout
	}
}

// usage
err := c.StartLogProducer(ctx, WithLogProducerTimeout(10*time.Second))
if err != nil {
	// do something with err
}
```

If no parameter is passed a default timeout of 5 seconds will be used. Values below 5 seconds and above 60 seconds will
be coerced to these boundary values.

## Listening to errors

When log producer fails to start within given timeout (causing a context deadline) or there's an error returned while closing the reader it will no longer panic, but instead will return an error over a channel. You can listen to it using `DockerContainer.GetLogProducerErrorChannel()` method:
```golang
func (c *DockerContainer) GetLogProducerErrorChannel() <-chan error {
	return c.producerError
}
```

This allows you to, for example, retry restarting log producer if it fails to start the first time. For example:

```golang
// start log producer normally
err = container.StartLogProducer(ctx, WithLogProducerTimeout(10*time.Second))

// listen to errors in a detached goroutine
go func(done chan struct{}, timeout time.Duration, retryLimit int) {
	for {
		select {
		case logErr := <-container.GetLogProducerErrorChannel():
			if logErr != nil {
				// do something with error
				// for example, retry starting log producer 
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