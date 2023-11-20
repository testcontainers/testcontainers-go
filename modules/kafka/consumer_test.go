package kafka

import (
	"testing"

	"github.com/IBM/sarama"
)

// TestKafkaConsumer is a test consumer for Kafka
type TestKafkaConsumer struct {
	t       *testing.T
	ready   chan bool
	done    chan bool
	cancel  chan bool
	message *sarama.ConsumerMessage
}

func NewTestKafkaConsumer(t *testing.T) (*TestKafkaConsumer, <-chan bool, <-chan bool, func()) {
	kc := &TestKafkaConsumer{
		t:      t,
		ready:  make(chan bool, 1),
		done:   make(chan bool, 1),
		cancel: make(chan bool, 1),
	}
	return kc, kc.ready, kc.done, func() {
		kc.cancel <- true
	}
}

func (k *TestKafkaConsumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (k *TestKafkaConsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim is called by the Kafka client library when a message is received
func (k *TestKafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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
