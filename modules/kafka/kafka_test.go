package kafka

import (
	"context"
	"strings"
	"testing"

	"github.com/IBM/sarama"

	"github.com/testcontainers/testcontainers-go"
)

func TestKafka(t *testing.T) {
	topic := "some-topic"

	ctx := context.Background()

	kafkaContainer, err := RunContainer(ctx, testcontainers.WithImage("confluentinc/cp-kafka:7.3.3"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// getBrokers {
	brokers, err := kafkaContainer.Brokers(ctx)
	// }
	if err != nil {
		t.Fatal(err)
	}

	config := sarama.NewConfig()
	client, err := sarama.NewConsumerGroup(brokers, "groupName", config)
	if err != nil {
		t.Fatal(err)
	}

	consumer, ready, done, cancel := NewTestKafkaConsumer(t)
	go func() {
		if err := client.Consume(context.Background(), []string{topic}, consumer); err != nil {
			cancel()
		}
	}()

	// wait for the consumer to be ready
	<-ready

	// perform assertions

	// set config to true because successfully delivered messages will be returned on the Successes channel
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

	if !strings.EqualFold(string(consumer.message.Key), "key") {
		t.Fatalf("expected key to be %s, got %s", "key", string(consumer.message.Key))
	}
	if !strings.EqualFold(string(consumer.message.Value), "value") {
		t.Fatalf("expected value to be %s, got %s", "value", string(consumer.message.Value))
	}
}
