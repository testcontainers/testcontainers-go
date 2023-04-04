package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStartContainer(t *testing.T) {
	topic := "some-topic"

	container, err := StartContainer(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	brokers, err := container.Brokers(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	config := sarama.NewConfig()
	client, err := sarama.NewConsumerGroup(brokers, "groupName", config)
	if err != nil {
		t.Fatal(err)
	}

	consumer, ready, done, cancel := newKafkaConsumer(t)
	go func() {
		if err := client.Consume(context.Background(), []string{topic}, consumer); err != nil {
			t.Fatal(err)
		}
	}()

	<-ready

	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		cancel()
		t.Fatal(err)
	}

	if _, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder("value"),
	}); err != nil {
		cancel()
		t.Fatal(err)
	}

	<-done

	assert.Equal(t, "key", string(consumer.message.Key))
	assert.Equal(t, "value", string(consumer.message.Value))
}

type kafkaConsumer struct {
	t       *testing.T
	ready   chan bool
	done    chan bool
	cancel  chan bool
	message *sarama.ConsumerMessage
}

func newKafkaConsumer(t *testing.T) (consumer *kafkaConsumer, ready <-chan bool, done <-chan bool, cancel func()) {
	kc := &kafkaConsumer{
		t:      t,
		ready:  make(chan bool, 1),
		done:   make(chan bool, 1),
		cancel: make(chan bool, 1),
	}
	return kc, kc.ready, kc.done, func() {
		kc.cancel <- true
	}
}

func (k *kafkaConsumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (k *kafkaConsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (k *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	k.ready <- true
	for {
		select {
		case message := <-claim.Messages():
			k.message = message
			session.MarkMessage(message, "")
			k.done <- true

		case <-k.cancel:
			return nil

		case <-session.Context().Done():
			return nil
		}
	}
}
