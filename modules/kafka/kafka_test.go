package kafka_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestKafka_Basic(t *testing.T) {
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
}

func TestKafka_invalidVersion(t *testing.T) {
	ctx := context.Background()

	ctr, err := kafka.Run(ctx, "confluentinc/confluent-local:6.3.3", kafka.WithClusterID("kraftCluster"))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}

func TestKafka_networkConnectivity(t *testing.T) {
	ctx := context.Background()
	var err error

	const (
		// config
		topic_in  = "topic_in"
		topic_out = "topic_out"

		address = "kafka:9092"

		// test data
		key      = "wow"
		text_msg = "test-input-external"
	)

	Network, err := network.New(ctx)
	require.NoError(t, err)

	// kafkaWithListener {
	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/confluent-local:7.6.1",
		kafka.WithClusterID("test-cluster"),
		network.WithNetwork([]string{"kafka"}, Network),
		kafka.WithListener(kafka.Listener{
			Name: "BROKER",
			Host: "kafka",
			Port: "9092",
		}),
	)
	// }
	testcontainers.CleanupContainer(t, kafkaContainer)
	require.NoError(t, err)

	kcat, err := runKcatContainer(ctx, Network.Name, "/tmp/msgs.txt")
	testcontainers.CleanupContainer(t, kcat)
	require.NoError(t, err)

	// 4. Copy message to kcat
	err = kcat.SaveFile(ctx, "Message produced by kcat")
	require.NoError(t, err)

	brokers, err := kafkaContainer.Brokers(context.TODO())
	require.NoError(t, err)

	// err = createTopics(brokers, []string{topic_in, topic_out})
	err = kcat.CreateTopic(ctx, address, topic_in)
	require.NoError(t, err)

	err = kcat.CreateTopic(ctx, address, topic_out)
	require.NoError(t, err)

	// perform assertions

	// set config to true because successfully delivered messages will be returned on the Successes channel
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Consumer.MaxWaitTime = 2 * time.Second

	producer, err := sarama.NewSyncProducer(brokers, config)
	require.NoError(t, err)

	// Act

	// External write
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic_in,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(text_msg),
	})
	require.NoError(t, err)

	// Internal read
	_, stdout, err := kcat.Exec(ctx, []string{"kcat", "-b", address, "-C", "-t", topic_in, "-c", "1"})
	require.NoError(t, err)

	out, err := io.ReadAll(stdout)
	require.NoError(t, err)
	require.Contains(t, string(out), text_msg)

	// Internal write
	tempfile := "/tmp/msgs.txt"

	err = kcat.CopyToContainer(ctx, []byte(out), tempfile, 700)
	require.NoError(t, err)

	_, _, err = kcat.Exec(ctx, []string{"kcat", "-b", address, "-t", topic_out, "-P", "-l", tempfile})
	require.NoError(t, err)

	// External read
	consumer := NewTestConsumer(t)

	client, err := sarama.NewConsumerGroup(brokers, "groupName", config)
	require.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	sCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(sCtx, []string{topic_out}, &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				require.NoError(t, err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	t.Log("Sarama consumer up and running!...")

	cancel()
	wg.Wait()
	err = client.Close()
	require.NoError(t, err)

	msgs := consumer.messages
	require.Len(t, msgs, 1)

	require.Contains(t, string(msgs[0].Value), text_msg)
	require.Contains(t, string(msgs[0].Key), key)
}

func TestKafka_withListener(t *testing.T) {
	ctx := context.Background()

	// 1. Create network
	rpNetwork, err := network.New(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := rpNetwork.Remove(ctx)
		require.NoError(t, err)
	})

	// 2. Start Kafka ctr
	// withListenerRP {
	ctr, err := kafka.Run(ctx,
		"confluentinc/confluent-local:7.6.1",
		network.WithNetwork([]string{"kafka"}, rpNetwork),
		kafka.WithListener(kafka.Listener{
			Name: "BROKER",
			Host: "kafka",
			Port: "9092",
		}),
	)
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// 3. Start KCat container
	// withListenerKcat {
	kcat, err := runKcatContainer(ctx, rpNetwork.Name, "/tmp/msgs.txt")
	// }
	testcontainers.CleanupContainer(t, kcat)
	require.NoError(t, err)

	// 4. Copy message to kcat
	err = kcat.SaveFile(ctx, "Message produced by kcat")
	require.NoError(t, err)

	// 5. Produce message to Kafka
	// withListenerExec {
	err = kcat.ProduceMessageFromFile(ctx, "kafka:9092", "msgs")
	// }

	require.NoError(t, err)

	// 6. Consume message from Kafka
	// 7. Read Message from stdout
	out, err := kcat.ConsumeMessage(ctx, "kafka:9092", "msgs")
	require.NoError(t, err)
	require.Contains(t, string(out), "Message produced by kcat")
}

func TestKafka_restProxyService(t *testing.T) {
	// TODO: test kafka rest proxy service
}

func TestKafka_listenersValidation(t *testing.T) {
	t.Run("reserved-listener/port-9093", func(t *testing.T) {
		runWithError(t, kafka.Listener{
			Name: "BROKER",
			Host: "kafka",
			Port: "9093",
		})
	})

	t.Run("reserved-listener/port-9094", func(t *testing.T) {
		runWithError(t, kafka.Listener{
			Name: "BROKER",
			Host: "kafka",
			Port: "9094",
		})
	})

	t.Run("reserved-listener/controller-duplicated", func(t *testing.T) {
		runWithError(t, kafka.Listener{
			Name: "  cOnTrOller   ",
			Host: "kafka",
			Port: "9092",
		})
	})

	t.Run("reserved-listener/plaintext-duplicated", func(t *testing.T) {
		runWithError(t, kafka.Listener{
			Name: "plaintext",
			Host: "kafka",
			Port: "9092",
		})
	})

	t.Run("duplicated-ports", func(t *testing.T) {
		runWithError(t, kafka.Listener{
			Name: "test",
			Host: "kafka",
			Port: "9092",
		},
			kafka.Listener{
				Name: "test2",
				Host: "kafka",
				Port: "9092",
			},
		)
	})

	t.Run("duplicated-names", func(t *testing.T) {
		runWithError(t, kafka.Listener{
			Name: "test",
			Host: "kafka",
			Port: "9092",
		},
			kafka.Listener{
				Name: "test",
				Host: "kafka",
				Port: "9095",
			},
		)
	})
}

// runWithError runs the Kafka container with the provided listeners and expects an error
func runWithError(t *testing.T, listeners ...kafka.Listener) {
	t.Helper()

	c, err := kafka.Run(context.Background(),
		"confluentinc/confluent-local:7.6.1",
		kafka.WithClusterID("test-cluster"),
		kafka.WithListener(listeners...),
	)
	require.Error(t, err)
	require.Nil(t, c)
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
