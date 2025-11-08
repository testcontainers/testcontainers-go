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
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	gracefulShutdownSupportingKafkaImages = []string{
		"apache/kafka:3.9.1",
		"apache/kafka:4.0.1",
		"apache/kafka-native:3.9.1",
		"apache/kafka-native:4.0.1",
	}

	supportedKafkaImages = append(gracefulShutdownSupportingKafkaImages,
		"confluentinc/confluent-local:7.4.0",
		"confluentinc/confluent-local:7.5.0",
	)
)

func TestKafka(t *testing.T) {
	for _, image := range supportedKafkaImages {
		t.Run(image, func(t *testing.T) {
			topic := "some-topic"

			ctx := context.Background()

			kafkaContainer, err := kafka.Run(ctx, image, kafka.WithClusterID("kraftCluster"))
			testcontainers.CleanupContainer(t, kafkaContainer, testcontainers.StopTimeout(0))
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
			defer func() {
				err := client.Close()
				require.NoError(t, err)
			}()

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
			defer func() {
				err := producer.Close()
				require.NoError(t, err)
			}()

			_, _, err = producer.SendMessage(&sarama.ProducerMessage{
				Topic: topic,
				Key:   sarama.StringEncoder("key"),
				Value: sarama.StringEncoder("value"),
			})
			require.NoError(t, err)

			<-done

			require.Truef(t, strings.EqualFold(string(consumer.message.Key), "key"), "expected key to be %s, got %s", "key", string(consumer.message.Key))
			require.Truef(t, strings.EqualFold(string(consumer.message.Value), "value"), "expected value to be %s, got %s", "value", string(consumer.message.Value))
		})
	}
}

func TestKafkaGracefulShutdown(t *testing.T) {
	for _, image := range gracefulShutdownSupportingKafkaImages {
		t.Run(image, func(t *testing.T) {
			ctx := context.Background()
			kafkaContainer, err := kafka.Run(ctx, image)
			testcontainers.CleanupContainer(t, kafkaContainer, testcontainers.StopTimeout(0))
			require.NoError(t, err)

			done := make(chan struct{})
			var stopErr error
			go func() {
				stopTimeout := 120 * time.Second
				stopErr = kafkaContainer.Stop(ctx, &stopTimeout)
				close(done)
			}()
			gracefulShutdownTimeout := 60 * time.Second
			select {
			case <-done:
				require.NoError(t, stopErr)
			case <-time.After(gracefulShutdownTimeout):
				require.Failf(t, "Kafka did not gracefully exit", "Kafka did not gracefully exit in %v", gracefulShutdownTimeout)
			}
		})
	}
}

func TestKafkaLocalhostListener(t *testing.T) {
	ctx := context.Background()
	kafkaContainer, err := kafka.Run(ctx, "apache/kafka:4.0.1",
		testcontainers.WithWaitStrategy(
			wait.NewExecStrategy([]string{
				"/opt/kafka/bin/kafka-topics.sh",
				"--bootstrap-server",
				"localhost:9095",
				"--list",
			}).
				WithExitCode(0).
				WithPollInterval(2*time.Second).
				WithStartupTimeout(120*time.Second)))
	testcontainers.CleanupContainer(t, kafkaContainer, testcontainers.StopTimeout(0))
	require.NoError(t, err)
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
