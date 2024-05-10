package kafka

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/docker/go-connections/nat"
	"golang.org/x/mod/semver"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const publicPort = nat.Port("9093/tcp")
const (
	starterScript = "/usr/sbin/testcontainers_start.sh"

	// starterScript {
	starterScriptContent = `#!/bin/bash
source /etc/confluent/docker/bash-config
export KAFKA_ADVERTISED_LISTENERS=%s
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
	ClusterID string
	Listeners KafkaListener
}

type KafkaListener struct {
	Name string
	Ip   string
	Port string
}

// RunContainer creates an instance of the Kafka container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "confluentinc/confluent-local:7.5.0",
		ExposedPorts: []string{string(publicPort)},
		Env: map[string]string{
			// envVars {
			"KAFKA_LISTENERS":                                "EXTERNAL://0.0.0.0:9093,INTERNAL://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094",
			"KAFKA_REST_BOOTSTRAP_SERVERS":                   "EXTERNAL://0.0.0.0:9093,INTERNAL://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "INTERNAL:PLAINTEXT,EXTERNAL:PLAINTEXT,CONTROLLER:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":               "INTERNAL",
			"KAFKA_BROKER_ID":                                "1",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_OFFSETS_TOPIC_NUM_PARTITIONS":             "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_LOG_FLUSH_INTERVAL_MESSAGES":              fmt.Sprintf("%d", math.MaxInt),
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
			"KAFKA_NODE_ID":                                  "1",
			"KAFKA_PROCESS_ROLES":                            "broker,controller",
			"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
			// }
		},
		Entrypoint: []string{"sh"},
		// this CMD will wait for the starter script to be copied into the container and then execute it
		Cmd: []string{"-c", "while [ ! -f " + starterScript + " ]; do sleep 0.1; done; bash " + starterScript},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&settings)
		}
		opt.Customize(&genericContainerReq)
	}

	trimListeners(settings.Listeners)
	if err := validateListeners(settings.Listeners); err != nil {
		return nil, fmt.Errorf("listeners validation: %w", err)
	}

	// apply envs for listeners
	envChange := editEnvsForListeners(settings.Listeners)
	for key, item := range envChange {
		genericContainerReq.Env[key] = item
	}

	genericContainerReq.ContainerRequest.LifecycleHooks =
		[]testcontainers.ContainerLifecycleHooks{
			{
				PostStarts: []testcontainers.ContainerHook{
					// 1. copy the starter script into the container
					func(ctx context.Context, c testcontainers.Container) error {
						if len(settings.Listeners) == 0 {
							defaultInternal, err := internalListener(ctx, c)
							if err != nil {
								return fmt.Errorf("can't create default internal listener: %w", err)
							}
							settings.Listeners = append(settings.Listeners, defaultInternal)
						}

						defaultExternal, err := externalListener(ctx, c)
						if err != nil {
							return fmt.Errorf("can't create default external listener: %w", err)
						}

						settings.Listeners = append(settings.Listeners, defaultExternal)

						var advertised []string
						for _, item := range settings.Listeners {
							advertised = append(advertised, fmt.Sprintf("%s://%s:%s", item.Name, item.Ip, item.Port))
						}

						scriptContent := fmt.Sprintf(starterScriptContent, strings.Join(advertised, ","))

						return c.CopyToContainer(ctx, []byte(scriptContent), starterScript, 0o755)
					},
					// 2. wait for the Kafka server to be ready
					func(ctx context.Context, c testcontainers.Container) error {
						return wait.ForLog(".*Transitioning from RECOVERY to RUNNING.*").AsRegexp().WaitUntilReady(ctx, c)
					},
				},
			},
		}

	err := validateKRaftVersion(genericContainerReq.Image)
	if err != nil {
		return nil, err
	}

	clusterID := genericContainerReq.Env["CLUSTER_ID"]

	configureControllerQuorumVoters(&genericContainerReq)

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &KafkaContainer{Container: container, ClusterID: clusterID}, nil
}

func trimListeners(listeners []KafkaListener) {
	for i := 0; i < len(listeners); i++ {
		listeners[i].Name = strings.ToUpper(strings.Trim(listeners[i].Name, " "))
		listeners[i].Ip = strings.Trim(listeners[i].Ip, " ")
		listeners[i].Port = strings.Trim(listeners[i].Port, " ")
	}
}

func validateListeners(listeners []KafkaListener) error {
	var ports map[string]bool = make(map[string]bool, len(listeners)+2)
	var names map[string]bool = make(map[string]bool, len(listeners)+2)

	// check for default listeners
	ports["9094"] = true
	ports["9093"] = true

	// check for default listeners
	names["CONTROLLER"] = true
	names["EXTERNAL"] = true

	for _, item := range listeners {
		if names[item.Name] {
			return fmt.Errorf("duplicate of listener name: %s", item.Name)
		}
		names[item.Name] = true

		if ports[item.Port] {
			return fmt.Errorf("duplicate of listener port: %s", item.Port)
		}
		ports[item.Port] = true
	}

	return nil
}

func editEnvsForListeners(listeners []KafkaListener) map[string]string {
	if len(listeners) == 0 {
		// no change
		return map[string]string{}
	}

	envs := map[string]string{
		"KAFKA_LISTENERS":                      "CONTROLLER://0.0.0.0:9094, EXTERNAL://0.0.0.0:9093",
		"KAFKA_REST_BOOTSTRAP_SERVERS":         "CONTROLLER://0.0.0.0:9094, EXTERNAL://0.0.0.0:9093",
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "CONTROLLER:PLAINTEXT, EXTERNAL:PLAINTEXT",
	}

	// expect first listener has common network between kafka instances
	envs["KAFKA_INTER_BROKER_LISTENER_NAME"] = listeners[0].Name

	// expect small number of listeners, so joins is okay
	for _, item := range listeners {
		envs["KAFKA_LISTENERS"] = strings.Join(
			[]string{
				envs["KAFKA_LISTENERS"],
				fmt.Sprintf("%s://0.0.0.0:%s", item.Name, item.Port),
			},
			",",
		)

		envs["KAFKA_REST_BOOTSTRAP_SERVERS"] = envs["KAFKA_LISTENERS"]

		envs["KAFKA_LISTENER_SECURITY_PROTOCOL_MAP"] = strings.Join(
			[]string{
				envs["KAFKA_LISTENER_SECURITY_PROTOCOL_MAP"],
				item.Name + ":" + "PLAINTEXT",
			},
			",",
		)
	}

	return envs
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
	// }
}

// validateKRaftVersion validates if the image version is compatible with KRaft mode,
// which is available since version 7.0.0.
func validateKRaftVersion(fqName string) error {
	if fqName == "" {
		return fmt.Errorf("image cannot be empty")
	}

	image := fqName[:strings.LastIndex(fqName, ":")]
	version := fqName[strings.LastIndex(fqName, ":")+1:]

	if !strings.EqualFold(image, "confluentinc/confluent-local") {
		// do not validate if the image is not the official one.
		// not raising an error here, letting the image to start and
		// eventually evaluate an error if it exists.
		return nil
	}

	// semver requires the version to start with a "v"
	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	if semver.Compare(version, "v7.4.0") < 0 { // version < v7.4.0
		return fmt.Errorf("version=%s. KRaft mode is only available since version 7.4.0", version)
	}

	return nil
}
