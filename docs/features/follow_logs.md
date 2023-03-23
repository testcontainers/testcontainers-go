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
using `c.StopLogProducer()`. For a particular container, only one `LogProducer` can be active at time
