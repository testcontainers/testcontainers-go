package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"testing"
)

func TestPubsub(t *testing.T) {
	ctx := context.Background()

	container, err := setupPubsub(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	t.Setenv("PUBSUB_EMULATOR_HOST", container.URI)
	client, err := pubsub.NewClient(ctx, "my-project-id")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	topic, err := client.CreateTopic(ctx, "greetings")
	if err != nil {
		t.Fatal(err)
	}
	subscription, err := client.CreateSubscription(ctx, "subscription",
		pubsub.SubscriptionConfig{Topic: topic})
	if err != nil {
		t.Fatal(err)
	}
	result := topic.Publish(ctx, &pubsub.Message{Data: []byte("Hello World")})
	_, err = result.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// perform assertions
	var data []byte
	cctx, cancel := context.WithCancel(ctx)
	err = subscription.Receive(cctx, func(ctx context.Context, m *pubsub.Message) {
		data = m.Data
		m.Ack()
		defer cancel()
	})
	if string(data) != "Hello World" {
		t.Fatalf("Expected value %s. Got %s.", "Hello World", data)
	}
}
