package kafka_test

import (
	"context"
	"strings"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func TestKafka(t *testing.T) {
	topic := "some-topic"

	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithClusterID("kraftCluster"))
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

	_, err = kafkaContainer.BrokersTLS(ctx)
	require.Error(t, err)

	_, err = kafkaContainer.TLSConfig()
	require.Error(t, err)
}

// TestKafka_withTLS checks that a client produces and consumes messages over the SSL listener,
// that a client not trusting the generated CA is rejected, and that PLAINTEXT keeps working.
func TestKafka_withTLS(t *testing.T) {
	topic := "some-tls-topic"

	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithTLS())
	testcontainers.CleanupContainer(t, kafkaContainer)
	require.NoError(t, err)

	// getBrokersTLS {
	brokers, err := kafkaContainer.BrokersTLS(ctx)
	require.NoError(t, err)

	tlsConfig, err := kafkaContainer.TLSConfig()
	require.NoError(t, err)

	config := sarama.NewConfig()
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = tlsConfig
	// }

	client, err := sarama.NewConsumerGroup(brokers, "groupName", config)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	consumer, ready, done, cancel := NewTestKafkaConsumer(t)
	defer cancel()
	go func() {
		if err := client.Consume(context.Background(), []string{topic}, consumer); err != nil {
			cancel()
		}
	}()

	// wait for the consumer to be ready
	<-ready

	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, producer.Close()) })

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder("value"),
	})
	require.NoError(t, err)

	<-done

	require.Equal(t, "key", string(consumer.message.Key))
	require.Equal(t, "value", string(consumer.message.Value))

	// a client that does not trust the generated CA is rejected
	untrustedConfig := sarama.NewConfig()
	untrustedConfig.Net.TLS.Enable = true
	_, err = sarama.NewClient(brokers, untrustedConfig)
	require.Error(t, err)

	// the PLAINTEXT listener keeps working next to the SSL one
	plainBrokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)
	require.NotEqual(t, brokers, plainBrokers)

	plainClient, err := sarama.NewClient(plainBrokers, sarama.NewConfig())
	require.NoError(t, err)
	require.NoError(t, plainClient.Close())
}

// TestKafka_withTLSInvalidListener checks that Run fails when the user defines
// an SSL listener on a port other than the one exposed by the module.
func TestKafka_withTLSInvalidListener(t *testing.T) {
	ctx := context.Background()

	ctr, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0",
		kafka.WithTLS(),
		testcontainers.WithEnv(map[string]string{
			"KAFKA_LISTENERS": "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094,SSL://0.0.0.0:9096",
		}),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.ErrorContains(t, err, "requires it to listen on port 9095")
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
