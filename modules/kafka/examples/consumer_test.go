package examples_test

import (
	"sync"
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
	t.Helper()
	kc := &TestKafkaConsumer{
		t:      t,
		ready:  make(chan bool, 1),
		done:   make(chan bool, 1),
		cancel: make(chan bool, 1),
	}
	var once sync.Once
	return kc, kc.ready, kc.done, func() {
		once.Do(func() { close(kc.cancel) })
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
	select {
	case k.ready <- true:
	case <-session.Context().Done():
		return nil
	case <-k.cancel:
		return nil
	}
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			k.message = message
			session.MarkMessage(message, "")
			select {
			case k.done <- true:
			case <-session.Context().Done():
				return nil
			case <-k.cancel:
				return nil
			}

		case <-k.cancel:
			return nil

		case <-session.Context().Done():
			return nil
		}
	}
}
