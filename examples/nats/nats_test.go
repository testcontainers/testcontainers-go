package nats

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
)

func TestNats(t *testing.T) {
	ctx := context.Background()

	container, err := runContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	nc, err := nats.Connect(container.URI)
	if err != nil {
		t.Fatalf("failed to connect to nats: %s", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("failed to create jetstream context: %s", err)
	}

	// add stream to nats
	if _, err = js.AddStream(&nats.StreamConfig{
		Name:     "hello",
		Subjects: []string{"hello"},
	}); err != nil {
		t.Fatalf("failed to add stream: %s", err)
	}

	// add subscriber to nats
	sub, err := js.SubscribeSync("hello", nats.Durable("worker"))
	if err != nil {
		t.Fatalf("failed to subscribe to hello: %s", err)
	}

	// publish a message to nats
	if _, err = js.Publish("hello", []byte("hello")); err != nil {
		t.Fatalf("failed to publish hello: %s", err)
	}

	// wait for the message to be received
	msg, err := sub.NextMsgWithContext(ctx)
	if err != nil {
		t.Fatalf("failed to get message: %s", err)
	}

	if string(msg.Data) != "hello" {
		t.Fatalf("expected message to be 'hello', got '%s'", msg.Data)
	}
}

