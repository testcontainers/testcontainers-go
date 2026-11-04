package kafka_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func TestKafka(t *testing.T) {
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

	require.NotEmpty(t, brokers)

	_, err = kafkaContainer.BrokersTLS(ctx)
	require.Error(t, err)

	_, err = kafkaContainer.TLSConfig()
	require.Error(t, err)
}

// TestKafka_withTLS checks that TLS and PLAINTEXT broker endpoints and TLS configuration are available.
func TestKafka_withTLS(t *testing.T) {
	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithTLS())
	testcontainers.CleanupContainer(t, kafkaContainer)
	require.NoError(t, err)

	// getBrokersTLS {
	brokers, err := kafkaContainer.BrokersTLS(ctx)
	require.NoError(t, err)

	tlsConfig, err := kafkaContainer.TLSConfig()
	require.NoError(t, err)
	require.NotNil(t, tlsConfig)
	// }

	plainBrokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)
	require.NotEqual(t, brokers, plainBrokers)
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
