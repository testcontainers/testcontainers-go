package kafka_test

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
	t.Helper()
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

// Consumer represents a Sarama consumer group consumer
type TestConsumer struct {
	t        *testing.T
	ready    chan bool
	messages []*sarama.ConsumerMessage
}

func NewTestConsumer(t *testing.T) TestConsumer {
	t.Helper()

	return TestConsumer{
		t:     t,
		ready: make(chan bool),
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *TestConsumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *TestConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (c *TestConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				c.t.Log("message channel was closed")
				return nil
			}
			c.t.Logf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			session.MarkMessage(message, "")

			// Store the message to be consumed later
			c.messages = append(c.messages, message)

		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/IBM/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}
