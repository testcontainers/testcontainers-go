package kafka

import (
	"context"
	"fmt"
	"math"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const publicPort = nat.Port("9093/tcp")
const (
	starterScript = "/usr/sbin/testcontainers_start.sh"

	// starterScript {
	starterScriptContent = `#!/bin/bash
source /etc/confluent/docker/bash-config
export KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://%s:%d,BROKER://%s:9092
echo Starting Kafka KRaft mode
sed -i '/KAFKA_ZOOKEEPER_CONNECT/d' /etc/confluent/docker/configure
echo 'kafka-storage format --ignore-formatted -t "$(kafka-storage random-uuid)" -c /etc/kafka/kafka.properties' >> /etc/confluent/docker/configure
echo '' > /etc/confluent/docker/ensure
/etc/confluent/docker/configure
/etc/confluent/docker/launch`
	// }
)

// KafkaContainer represents the Kafka container type used in the module
type KafkaContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Kafka container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-kafka:7.3.3",
		ExposedPorts: []string{string(publicPort)},
		Env: map[string]string{
			// envVars {
			"KAFKA_LISTENERS":                                "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":               "BROKER",
			"KAFKA_BROKER_ID":                                "1",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_OFFSETS_TOPIC_NUM_PARTITIONS":             "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_LOG_FLUSH_INTERVAL_MESSAGES":              fmt.Sprintf("%d", math.MaxInt64),
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
			"KAFKA_NODE_ID":                                  "1",
			"KAFKA_PROCESS_ROLES":                            "broker,controller",
			"KAFKA_CONTROLLER_QUORUM_VOTERS":                 "1@localhost:9094",
			"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
			// }
		},
		Entrypoint: []string{"sh"},
		// this CMD will wait for the starter script to be copied into the container and then execute it
		Cmd: []string{"-c", "while [ ! -f " + starterScript + " ]; do sleep 0.1; done; bash " + starterScript},
		LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
			{
				PostStarts: []testcontainers.ContainerHook{
					// 1. copy the starter script into the container
					func(ctx context.Context, c testcontainers.Container) error {
						host, err := c.Host(ctx)
						if err != nil {
							return err
						}

						port, err := c.MappedPort(ctx, publicPort)
						if err != nil {
							return err
						}

						scriptContent := fmt.Sprintf(starterScriptContent, host, port.Int(), host)

						return c.CopyToContainer(ctx, []byte(scriptContent), starterScript, 0o755)
					},
					// 2. wait for the Kafka server to be ready
					func(ctx context.Context, c testcontainers.Container) error {
						return wait.ForLog("Kafka Server started").WaitUntilReady(ctx, c)
					},
				},
			},
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &KafkaContainer{Container: container}, nil
}

// Brokers retrieves the broker connection strings from Kafka with only one entry,
// defined by the exposed public port.
func (kc *KafkaContainer) Brokers(ctx context.Context) ([]string, error) {
	host, err := kc.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := kc.MappedPort(ctx, publicPort)
	if err != nil {
		return nil, err
	}

	return []string{fmt.Sprintf("%s:%d", host, port.Int())}, nil
}
