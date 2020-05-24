# Following Container Logs

If you wish to follow container logs, you can set up `LogConsumer`s.  The log
following functionality follows a producer-consumer model.  You will need to
explicitly start and stop the producer.  As logs are written to either `stdout`,
or `stderr` (`stdin` is not supported) they will be forwarded (produced) to any
associated `LogConsumer`s.  You can associate `LogConsumer`s with the
`.FollowOutput` function.

**Please note** if you start the producer you should always stop it explicitly.

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

err := c.StartLogProducer(ctx)
if err != nil {
	// do something with err
}

c.FollowOutput(&g)

// some stuff happens...

err = c.StopLogProducer()
if err != nil {
	// do something with err
}
```

