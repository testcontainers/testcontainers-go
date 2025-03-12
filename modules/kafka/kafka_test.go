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
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestKafka_invalidVersion(t *testing.T) {
	ctx := context.Background()

	ctr, err := kafka.Run(ctx, "confluentinc/confluent-local:6.3.3", kafka.WithClusterID("kraftCluster"))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}

const (
	testTopic = "some-topic"
	testValue = "kafka-message-value"
)

func TestKafka(t *testing.T) {
	ctx := context.Background()
	net, err := network.New(ctx)
	testcontainers.CleanupNetwork(t, net)
	require.NoError(t, err)

	kafkaContainer, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithClusterID("kraftCluster"), network.WithNetwork(nil, net))
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
		if err := client.Consume(context.Background(), []string{testTopic}, consumer); err != nil {
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
		Topic: testTopic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder(testValue),
	})
	require.NoError(t, err)

	<-done

	require.Truef(t, strings.EqualFold(string(consumer.message.Key), "key"), "expected key to be %s, got %s", "key", string(consumer.message.Key))
	require.Truef(t, strings.EqualFold(string(consumer.message.Value), testValue), "expected value to be %s, got %s", testValue, string(consumer.message.Value))

	t.Run("BrokersByHostDockerInternal", func(t *testing.T) {
		brokers, err := kafkaContainer.BrokersByHostDockerInternal(ctx)
		require.NoError(t, err)

		kcat, err := runKcatContainer(ctx, brokers, func(hc *container.HostConfig) {
			hc.ExtraHosts = append(hc.ExtraHosts, "host.docker.internal:host-gateway")
		}, nil)
		testcontainers.CleanupContainer(t, kcat)
		require.NoError(t, err)

		l, err := kcat.Logs(ctx)
		require.NoError(t, err)
		defer l.Close()

		assertKcatReadMsg(t, l)
	})
	t.Run("BrokersByContainerId", func(t *testing.T) {
		brokers, err := kafkaContainer.BrokersByContainerId(ctx)
		require.NoError(t, err)

		kcat, err := runKcatContainer(ctx, brokers, nil, []string{net.Name})
		testcontainers.CleanupContainer(t, kcat)
		require.NoError(t, err)

		l, err := kcat.Logs(ctx)
		require.NoError(t, err)
		defer l.Close()

		assertKcatReadMsg(t, l)
	})
	t.Run("BrokersByContainerName", func(t *testing.T) {
		brokers, err := kafkaContainer.BrokersByContainerName(ctx)
		require.NoError(t, err)

		kcat, err := runKcatContainer(ctx, brokers, nil, []string{net.Name})
		testcontainers.CleanupContainer(t, kcat)
		require.NoError(t, err)

		l, err := kcat.Logs(ctx)
		require.NoError(t, err)
		defer l.Close()

		assertKcatReadMsg(t, l)
	})
}

func runKcatContainer(ctx context.Context, brokers []string, hostMod func(*container.HostConfig), networks []string) (testcontainers.Container, error) {
	return testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:              "confluentinc/cp-kafkacat",
				Entrypoint:         []string{"kafkacat"},
				Cmd:                []string{"-b", strings.Join(brokers, ","), "-C", "-t", testTopic, "-c", "1"},
				WaitingFor:         wait.ForExit(),
				HostConfigModifier: hostMod,
				Networks:           networks,
			},
			Started: true,
		},
	)
}

func assertKcatReadMsg(t *testing.T, l io.Reader) {
	t.Helper()
	lb, err := io.ReadAll(l)
	require.NoError(t, err)

	readMsg := string(bytes.TrimSpace(lb))
	require.Truef(t, strings.EqualFold(readMsg, testValue), "expected value to be %s, got %s", testValue, readMsg)
}

const (
	// Internal listening port for Broker intercommunication
	brokerToBrokerPort = 9092
	// Mapped port for advertised listener of localhost:<mapped_port>.
	publicLocalhostPort = nat.Port("9093/tcp")
	// Mapped port for advertised listener of host.docker.internal:<mapped_port>
	publicDockerHostPort = nat.Port("19093/tcp")
	// Internal listening port for advertised listener of <container_name>:19094. This is not mapped to a random host port
	networkInternalContainerNamePort = 19094
	// Internal listening port for advertised listener of <container_id>:19095. This is not mapped to a random host port
	networkInternalContainerIdPort = 19095
)

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

	portLh, err := container.MappedPort(ctx, publicLocalhostPort)
	require.NoError(t, err)

	portDh, err := container.MappedPort(ctx, publicDockerHostPort)
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

	assert(fmt.Sprintf("CONTAINER_NAME://%s:%d", strings.Trim(inspect.Name, "/"), networkInternalContainerNamePort))
	assert(fmt.Sprintf("CONTAINER_NAME://%s", mustBrokers(container.BrokersByContainerName))) //nolint:perfsprint

	assert(fmt.Sprintf("CONTAINER_ID://%s:%d", inspect.Config.Hostname, networkInternalContainerIdPort))
	assert(fmt.Sprintf("CONTAINER_ID://%s", mustBrokers(container.BrokersByContainerId))) //nolint:perfsprint

	assert(fmt.Sprintf("BROKER://%s:%d", inspect.Config.Hostname, brokerToBrokerPort))
}
