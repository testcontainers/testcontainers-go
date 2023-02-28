package pulsar

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

func TestPulsar(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := startContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	pc, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               c.URI,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })

	consumer, err := pc.Subscribe(pulsar.ConsumerOptions{
		Topic:            "test-topic",
		SubscriptionName: "pulsar-test",
		Type:             pulsar.Exclusive,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { consumer.Close() })

	msgChan := make(chan []byte)
	go func() {
		msg, err := consumer.Receive(ctx)
		if err != nil {
			fmt.Println("failed to receive message", err)
			return
		}
		msgChan <- msg.Payload()
		consumer.Ack(msg)
	}()

	producer, err := pc.CreateProducer(pulsar.ProducerOptions{
		Topic: "test-topic",
	})
	if err != nil {
		t.Fatal(err)
	}

	producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: []byte("hello world"),
	})

	ticker := time.NewTicker(1 * time.Minute)
	select {
	case <-ticker.C:
		t.Fatal("did not receive message in time")
	case msg := <-msgChan:
		if string(msg) != "hello world" {
			t.Fatal("received unexpected message bytes")
		}
	}
}
