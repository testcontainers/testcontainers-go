# Pulsar

```go
package main

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPulsar(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	matchAdminResponse := func(r io.Reader) bool {
		respBytes, _ := io.ReadAll(r)
		resp := string(respBytes)
		return resp == `["standalone"]`
	}
	pulsarRequest := testcontainers.ContainerRequest{
		Image:        "docker.io/apachepulsar/pulsar:2.10.2",
		ExposedPorts: []string{"6650/tcp", "8080/tcp"},
		WaitingFor:   wait.ForHTTP("/admin/v2/clusters").WithPort("8080/tcp").WithResponseMatcher(matchAdminResponse),
		Cmd: []string{
			"/bin/bash",
			"-c",
			"/pulsar/bin/apply-config-from-env.py /pulsar/conf/standalone.conf && bin/pulsar standalone --no-functions-worker -nss",
		},
	}
	pulsarContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pulsarRequest,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pulsarContainer.Terminate(ctx) })

	pulsarContainer.StartLogProducer(ctx)
	defer pulsarContainer.StopLogProducer()
	lc := logConsumer{}
	pulsarContainer.FollowOutput(&lc)

	pulsarPort, err := pulsarContainer.MappedPort(ctx, "6650/tcp")
	if err != nil {
		t.Fatal(err)
	}

	pc, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               fmt.Sprintf("pulsar://127.0.0.1:%v", pulsarPort.Int()),
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

type logConsumer struct{}

func (lc *logConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}
```