package kafka

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/testcontainers/testcontainers-go/wait"
	"strconv"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

const (
	publicPort = nat.Port("9093/tcp")
)

// KafkaContainer represents the Kafka container type used in the module
type KafkaContainer struct {
	testcontainers.Container
	ClusterID string
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Kafka container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error) {
	return Run(ctx, "apache/kafka-native:3.8.0", opts...)
}

// Run creates an instance of the Kafka container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{string(publicPort)},
		WaitingFor: wait.ForLog("Transition from STARTING to STARTED (kafka." +
			"server.BrokerServer)").WithStartupTimeout(2 * time.Minute),
		Env: map[string]string{
			"KAFKA_LISTENERS":              "CONTROLLER://0.0.0.0:9094,PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092",
			"KAFKA_ADVERTISED_LISTENERS":   "PLAINTEXT://localhost:9093,BROKER://localhost:9092",
			"KAFKA_REST_BOOTSTRAP_SERVERS": "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "CONTROLLER" +
				":PLAINTEXT,SASL_PLAINTEXT:SASL_PLAINTEXT," +
				"PLAINTEXT:PLAINTEXT," +
				"BROKER:PLAINTEXT",
			"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
			"KAFKA_BROKER_ID":                                "1",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_OFFSETS_TOPIC_NUM_PARTITIONS":             "1",
			"KAFKA_INTER_BROKER_LISTENER_NAME":               "BROKER",
			"KAFKA_SASL_MECHANISM_INTER_BROKER_PROTOCOL":     "PLAIN",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_LOG_FLUSH_INTERVAL_MESSAGES":              "9223372036854775807", // math.MaxInt value
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
			"KAFKA_NODE_ID":                                  "1",
			"KAFKA_PROCESS_ROLES":                            "broker,controller",
			"KAFKA_CONTROLLER_QUORUM_VOTERS":                 "1@localhost:9094",
			"KAFKA_SASL_ENABLED_MECHANISMS":                  "PLAIN",
			"KAFKA_OPTS":                                     " ",
		},
		HostConfigModifier: func(config *container.HostConfig) {
			config.PortBindings = map[nat.Port][]nat.PortBinding{
				"9093/tcp": {
					{
						HostIP:   "0.0.0.0",
						HostPort: "9093",
					},
				},
			}
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	clusterID := genericContainerReq.Env["CLUSTER_ID"]
	configureControllerQuorumVoters(&genericContainerReq)

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &KafkaContainer{Container: container, ClusterID: clusterID}, nil
}

func WithClusterID(clusterID string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["CLUSTER_ID"] = clusterID

		return nil
	}
}

func WithHostPort(port int) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.HostConfigModifier = func(config *container.HostConfig) {
			config.PortBindings = map[nat.Port][]nat.PortBinding{
				"9093/tcp": {
					{
						HostIP:   "0.0.0.0",
						HostPort: strconv.Itoa(port),
					},
				},
			}
		}
		return nil
	}
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

// configureControllerQuorumVoters sets the quorum voters for the controller. For that, it will
// check if there are any network aliases defined for the container and use the first alias in the
// first network. Else, it will use localhost.
func configureControllerQuorumVoters(req *testcontainers.GenericContainerRequest) {
	if req.Env == nil {
		req.Env = map[string]string{}
	}

	if req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"] == "" {
		host := "localhost"
		if len(req.Networks) > 0 {
			nw := req.Networks[0]
			if len(req.NetworkAliases[nw]) > 0 {
				host = req.NetworkAliases[nw][0]
			}
		}

		req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"] = fmt.Sprintf("1@%s:9094", host)
	}
}
