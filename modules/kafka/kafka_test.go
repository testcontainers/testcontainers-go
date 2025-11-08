package kafka_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func testFor(t *testing.T, image string) {
	t.Helper()

	topic := "some-topic"

	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx, image, kafka.WithClusterID("kraftCluster"))
	testcontainers.CleanupContainer(t, kafkaContainer)
	require.NoError(t, err)

	assertAdvertisedListeners(t, kafkaContainer)

	require.Truef(t, strings.EqualFold(kafkaContainer.ClusterID, "kraftCluster"), "expected clusterID to be %s, got %s", "kraftCluster", kafkaContainer.ClusterID)

	// getBrokers {
	brokers, err := kafkaContainer.Brokers(ctx)
	// }
	require.NoError(t, err)

	config := sarama.NewConfig()
	client, err := sarama.NewConsumerGroup(brokers, "groupName", config)
	require.NoError(t, err)

	consumer, ready, done, cancel := NewTestKafkaConsumer(t)
	defer cancel()
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
	require.NoError(t, err)

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder("value"),
	})
	require.NoError(t, err)

	<-done

	require.Truef(t, strings.EqualFold(string(consumer.message.Key), "key"), "expected key to be %s, got %s", "key", string(consumer.message.Key))
	require.Truef(t, strings.EqualFold(string(consumer.message.Value), "value"), "expected value to be %s, got %s", "value", string(consumer.message.Value))
}

func TestKafka(t *testing.T) {
	testCases := []struct {
		name  string
		image string
	}{
		{
			name:  "confluentinc",
			image: "confluentinc/confluent-local:7.4.0",
		},
		{
			name:  "confluentinc",
			image: "confluentinc/confluent-local:7.5.0",
		},
		{
			name:  "apache native 4",
			image: "apache/kafka-native:4.0.1",
		},
		{
			name:  "apache not-native 4",
			image: "apache/kafka:4.0.1",
		},
		{
			name:  "apache native 3.9",
			image: "apache/kafka-native:3.9.1",
		},
		{
			name:  "apache not-native 3.9",
			image: "apache/kafka:3.9.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFor(t, tc.image)
		})
	}
}

func TestKafka_invalidVersion(t *testing.T) {
	ctx := context.Background()

	ctr, err := kafka.Run(ctx, "confluentinc/confluent-local:6.3.3", kafka.WithClusterID("kraftCluster"))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}

// assertAdvertisedListeners checks that the advertised listeners are set correctly:
// - The BROKER:// protocol is using the hostname of the Kafka container
func assertAdvertisedListeners(t *testing.T, container testcontainers.Container) {
	t.Helper()
	inspect, err := container.Inspect(context.Background())
	require.NoError(t, err)

	brokerURL := "BROKER://" + inspect.Config.Hostname + ":9092"

	ctx := context.Background()

	bs := testcontainers.RequireContainerExec(ctx, t, container, []string{"cat", "/usr/sbin/testcontainers_start.sh"})

	require.Containsf(t, bs, brokerURL, "expected advertised listeners to contain %s, got %s", brokerURL, bs)
}

func TestKafkaGracefulShutdown(t *testing.T) {
	testCases := []struct {
		name  string
		image string
	}{
		{
			name:  "apache native",
			image: "apache/kafka-native:4.0.1",
		},
		{
			name:  "apache not-native",
			image: "apache/kafka:4.0.1",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			kafkaContainer, err := kafka.Run(ctx, tc.image)
			testcontainers.CleanupContainer(t, kafkaContainer, testcontainers.StopTimeout(0))
			require.NoError(t, err)

			done := make(chan struct{})
			go func() {
				stopTimeout := 120 * time.Second
				_ = kafkaContainer.Stop(ctx, &stopTimeout)
				close(done)
			}()
			gracefulShutdownTimeout := 60 * time.Second
			select {
			case <-done:
			case <-time.After(gracefulShutdownTimeout):
				require.Failf(t, "Kafka did not gracefully exit", "Kafka did not gracefully exit in %v", gracefulShutdownTimeout)
			}
		})
	}
}
