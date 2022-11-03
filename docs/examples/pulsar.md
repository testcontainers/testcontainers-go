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

type pulsarContainer struct {
	testcontainers.Container
	URI string
}

func setupPulsar(ctx context.Context) (*pulsarContainer, error) {
	matchAdminResponse := func(r io.Reader) bool {
		respBytes, _ := io.ReadAll(r)
		resp := string(respBytes)
		return resp == `["standalone"]`
	}
	pulsarRequest := testcontainers.ContainerRequest{
		Image:        "docker.io/apachepulsar/pulsar:2.10.2",
		ExposedPorts: []string{"6650/tcp", "8080/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/admin/v2/clusters").WithPort("8080/tcp").WithResponseMatcher(matchAdminResponse),
			wait.ForLog("Successfully updated the policies on namespace public/default"),
		),
		Cmd: []string{
			"/bin/bash",
			"-c",
			"/pulsar/bin/apply-config-from-env.py /pulsar/conf/standalone.conf && bin/pulsar standalone --no-functions-worker -nss",
		},
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pulsarRequest,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	c.StartLogProducer(ctx)
	defer c.StopLogProducer()
	lc := logConsumer{}
	c.FollowOutput(&lc)

	pulsarPort, err := c.MappedPort(ctx, "6650/tcp")
	if err != nil {
		return nil, err
	}

	return &pulsarContainer{
		Container: c,
		URI:       fmt.Sprintf("pulsar://127.0.0.1:%v", pulsarPort.Int()),
	}, nil
}

func TestPulsar(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := setupPulsar(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Container.Terminate(ctx) })

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

type logConsumer struct{}

func (lc *logConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}
```