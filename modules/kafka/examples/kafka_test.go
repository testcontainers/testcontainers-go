package examples_test

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

func TestKafka(t *testing.T) {
	topic := "some-topic"

	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithClusterID("kraftCluster"))
	testcontainers.CleanupContainer(t, kafkaContainer)
	require.NoError(t, err)

	require.Truef(t, strings.EqualFold(kafkaContainer.ClusterID, "kraftCluster"), "expected clusterID to be %s, got %s", "kraftCluster", kafkaContainer.ClusterID)

	// getBrokers {
	brokers, err := kafkaContainer.Brokers(ctx)
	// }
	require.NoError(t, err)

	config := sarama.NewConfig()
	client, err := sarama.NewConsumerGroup(brokers, "groupName", config)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	consumer, ready, done, cancel := NewTestKafkaConsumer(t)
	defer cancel()
	consumeCtx, consumeCancel := context.WithTimeout(ctx, time.Minute)
	defer consumeCancel()
	go func() {
		if err := client.Consume(consumeCtx, []string{topic}, consumer); err != nil {
			cancel()
		}
	}()

	// wait for the consumer to be ready
	select {
	case <-ready:
	case <-consumeCtx.Done():
		t.Fatal("consumer did not become ready in time")
	}

	// perform assertions

	// set config to true because successfully delivered messages will be returned on the Successes channel
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

	select {
	case <-done:
	case <-consumeCtx.Done():
		t.Fatal("consumer did not receive the message in time")
	}

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
	consumeCtx, consumeCancel := context.WithTimeout(ctx, time.Minute)
	defer consumeCancel()
	go func() {
		if err := client.Consume(consumeCtx, []string{topic}, consumer); err != nil {
			cancel()
		}
	}()

	// wait for the consumer to be ready
	select {
	case <-ready:
	case <-consumeCtx.Done():
		t.Fatal("consumer did not become ready in time")
	}

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

	select {
	case <-done:
	case <-consumeCtx.Done():
		t.Fatal("consumer did not receive the message in time")
	}

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
