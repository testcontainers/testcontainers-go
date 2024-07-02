package kafka_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/IBM/sarama"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func TestKafka(t *testing.T) {
	topic := "some-topic"

	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithClusterID("kraftCluster"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertAdvertisedListeners(t, kafkaContainer)

	if !strings.EqualFold(kafkaContainer.ClusterID, "kraftCluster") {
		t.Fatalf("expected clusterID to be %s, got %s", "kraftCluster", kafkaContainer.ClusterID)
	}

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

func TestKafka_invalidVersion(t *testing.T) {
	ctx := context.Background()

	_, err := kafka.Run(ctx, "confluentinc/confluent-local:6.3.3", kafka.WithClusterID("kraftCluster"))
	if err == nil {
		t.Fatal(err)
	}
}

// assertAdvertisedListeners checks that the advertised listeners are set correctly:
// - The BROKER:// protocol is using the hostname of the Kafka container
func assertAdvertisedListeners(t *testing.T, container testcontainers.Container) {
	inspect, err := container.Inspect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	hostname := inspect.Config.Hostname

	code, r, err := container.Exec(context.Background(), []string{"cat", "/usr/sbin/testcontainers_start.sh"})
	if err != nil {
		t.Fatal(err)
	}

	if code != 0 {
		t.Fatalf("expected exit code to be 0, got %d", code)
	}

	bs, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(bs), "BROKER://"+hostname+":9092") {
		t.Fatalf("expected advertised listeners to contain %s, got %s", "BROKER://"+hostname+":9092", string(bs))
	}
}
