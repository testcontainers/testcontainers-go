package kafka_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/IBM/sarama"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestKafka(t *testing.T) {
	const (
		topic = "some-topic"
		value = "kafka-message-value"
	)

	ctx := context.Background()
	net, err := network.New(ctx)
	require.NoError(t, err)

	kafkaContainer, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithClusterID("kraftCluster"), network.WithNetwork(nil, net))
	testcontainers.CleanupContainer(t, kafkaContainer)
	require.NoError(t, err)

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		require.NoError(t, kafkaContainer.Terminate(ctx), "failed to terminate container: %v", err)
	})

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
		Value: sarama.StringEncoder(value),
	})
	require.NoError(t, err)

	<-done

	require.Truef(t, strings.EqualFold(string(consumer.message.Key), "key"), "expected key to be %s, got %s", "key", string(consumer.message.Key))
	require.Truef(t, strings.EqualFold(string(consumer.message.Value), value), "expected value to be %s, got %s", value, string(consumer.message.Value))

	assertBrokers := func(
		prefix string,
		getBrokers func(context.Context) ([]string, error),
		hostMod func(*container.HostConfig),
	) {
		t.Helper()

		brokers, err = getBrokers(ctx)
		require.NoError(t, err)

		t.Log(prefix, strings.Join(brokers, ","))

		kcat, err := testcontainers.GenericContainer(
			ctx,
			testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Image:              "confluentinc/cp-kafkacat",
					Entrypoint:         []string{"kafkacat"},
					Cmd:                []string{"-b", strings.Join(brokers, ","), "-C", "-t", topic, "-c", "1"},
					WaitingFor:         wait.ForExit(),
					HostConfigModifier: hostMod,
					Networks:           []string{net.Name},
				},
				Started: true,
			},
		)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, kcat.Terminate(ctx), "failed to terminate container")
		})

		l, err := kcat.Logs(ctx)
		require.NoError(t, err)

		lb, err := io.ReadAll(l)
		require.NoError(t, err)

		readMsg := string(bytes.TrimSpace(lb))
		require.Truef(t, strings.EqualFold(readMsg, value), "expected value to be %s, got %s", value, readMsg)
	}

	t.Run("BrokersByHostDockerInternal", func(t *testing.T) {
		assertBrokers("BrokersByHostDockerInternal: ", kafkaContainer.BrokersByHostDockerInternal, func(hc *container.HostConfig) {
			hc.ExtraHosts = append(hc.ExtraHosts, "host.docker.internal:host-gateway")
		})
	})
	t.Run("BrokersByContainerId", func(t *testing.T) {
		assertBrokers("BrokersByContainerId: ", kafkaContainer.BrokersByContainerId, nil)
	})
	t.Run("BrokersByContainerName", func(t *testing.T) {
		assertBrokers("BrokersByContainerName: ", kafkaContainer.BrokersByContainerName, nil)
	})
}

func TestKafka_invalidVersion(t *testing.T) {
	ctx := context.Background()

	ctr, err := kafka.Run(ctx, "confluentinc/confluent-local:6.3.3", kafka.WithClusterID("kraftCluster"))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}

// assertAdvertisedListeners checks that the advertised listeners are set correctly:
// - The LOCALHOST:// protocol is using the host of the Kafka container
// - The HOST_DOCKER_INTERNAL:// protocol is using hostname host.docker.internal
// - The CONTAINER_NAME:// protocol is using the container name of the Kafka container
// - The CONTAINER_ID:// protocol is using the container ID of the Kafka container
// - The BROKER:// protocol is using the hostname of the Kafka container
func assertAdvertisedListeners(t *testing.T, container *kafka.KafkaContainer) {
	t.Helper()
	ctx := context.Background()

	inspect, err := container.Inspect(ctx)
	require.NoError(t, err)

	portLh, err := container.MappedPort(ctx, kafka.PublicLocalhostPort)
	require.NoError(t, err)

	portDh, err := container.MappedPort(ctx, kafka.PublicDockerHostPort)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	bs := testcontainers.RequireContainerExec(ctx, t, container, []string{"cat", "/usr/sbin/testcontainers_start.sh"})

	assert := func(listener string) {
		t.Helper()
		require.Containsf(t, bs, listener, "expected advertised listeners to contain %s, got %s", listener, bs)
	}

	mustBrokers := func(fn func(context.Context) ([]string, error)) string {
		t.Helper()
		brokers, err := fn(ctx)
		require.NoError(t, err)
		require.Len(t, brokers, 1)
		return brokers[0]
	}

	assert(fmt.Sprintf("LOCALHOST://%s:%d", host, portLh.Int()))
	assert(fmt.Sprintf("LOCALHOST://%s", mustBrokers(container.Brokers))) //nolint:perfsprint

	assert(fmt.Sprintf("HOST_DOCKER_INTERNAL://host.docker.internal:%d", portDh.Int()))
	assert(fmt.Sprintf("HOST_DOCKER_INTERNAL://%s", mustBrokers(container.BrokersByHostDockerInternal))) //nolint:perfsprint

	assert(fmt.Sprintf("CONTAINER_NAME://%s:%d", strings.Trim(inspect.Name, "/"), kafka.NetworkInternalContainerNamePort))
	assert(fmt.Sprintf("CONTAINER_NAME://%s", mustBrokers(container.BrokersByContainerName))) //nolint:perfsprint

	assert(fmt.Sprintf("CONTAINER_ID://%s:%d", inspect.Config.Hostname, kafka.NetworkInternalContainerIdPort))
	assert(fmt.Sprintf("CONTAINER_ID://%s", mustBrokers(container.BrokersByContainerId))) //nolint:perfsprint

	assert(fmt.Sprintf("BROKER://%s:%d", inspect.Config.Hostname, kafka.BrokerToBrokerPort))
}
