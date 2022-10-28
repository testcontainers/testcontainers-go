# Pulsar

```go
package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPulsar(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g := NewGomegaWithT(t)

	matchAdminResponse := func(r io.Reader) bool {
		respBytes, _ := ioutil.ReadAll(r)
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
	g.Expect(err).ToNot(HaveOccurred())
	t.Cleanup(func() { pulsarContainer.Terminate(ctx) })

	pulsarContainer.StartLogProducer(ctx)
	defer pulsarContainer.StopLogProducer()
	lc := logConsumer{}
	pulsarContainer.FollowOutput(&lc)

	pulsarPort, err := pulsarContainer.MappedPort(ctx, "6650/tcp")
	g.Expect(err).ToNot(HaveOccurred())

	pc, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               fmt.Sprintf("pulsar://127.0.0.1:%v", pulsarPort.Int()),
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	g.Expect(err).ToNot(HaveOccurred())
	t.Cleanup(func() { pc.Close() })

	consumer, err := pc.Subscribe(pulsar.ConsumerOptions{
		Topic:            "test-topic",
		SubscriptionName: "pulsar-test",
		Type:             pulsar.Exclusive,
	})
	g.Expect(err).ToNot(HaveOccurred())
	t.Cleanup(func() { consumer.Close() })

	msgChan := make(chan []byte)
	go func() {
		msg, err := consumer.Receive(ctx)
		g.Expect(err).ToNot(HaveOccurred())
		msgChan <- msg.Payload()
		consumer.Ack(msg)
	}()

	producer, err := pc.CreateProducer(pulsar.ProducerOptions{
		Topic: "test-topic",
	})
	g.Expect(err).ToNot(HaveOccurred())

	producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: []byte("hello world"),
	})

	g.Eventually(msgChan).Should(Receive(Equal([]byte("hello world"))))
}

type logConsumer struct{}

func (lc *logConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}
```