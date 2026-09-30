package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	publicPort    = "9093/tcp"
	sslPortNumber = "9095"
	sslPort       = sslPortNumber + "/tcp"
)

const (
	starterScript = "/usr/sbin/testcontainers_start.sh"

	// starterScript {
	starterScriptContent = `#!/bin/bash
source /etc/confluent/docker/bash-config
export KAFKA_ADVERTISED_LISTENERS=%s,BROKER://%s:9092
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
	settings  *options
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Kafka container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error) {
	return Run(ctx, "confluentinc/confluent-local:7.5.0", opts...)
}

// Run creates an instance of the Kafka container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error) {
	if err := validateKRaftVersion(img); err != nil {
		return nil, err
	}

	// Process custom options to extract settings. The settings are shared with
	// the post-start hook, which generates the TLS certificates on the first start.
	settings := &options{}
	for _, opt := range opts {
		if opt, ok := opt.(Option); ok {
			if err := opt(settings); err != nil {
				return nil, fmt.Errorf("apply option: %w", err)
			}
		}
	}

	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 5+len(opts)+1)
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(string(publicPort)),
		testcontainers.WithEnv(map[string]string{
			// envVars {
			"KAFKA_LISTENERS":                                "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094",
			"KAFKA_REST_BOOTSTRAP_SERVERS":                   "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":               "BROKER",
			"KAFKA_BROKER_ID":                                "1",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_OFFSETS_TOPIC_NUM_PARTITIONS":             "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_LOG_FLUSH_INTERVAL_MESSAGES":              strconv.Itoa(math.MaxInt),
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
			"KAFKA_NODE_ID":                                  "1",
			"KAFKA_PROCESS_ROLES":                            "broker,controller",
			"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
			// }
		}),
		testcontainers.WithEntrypoint("sh"),
		// this CMD will wait for the starter script to be copied into the container and then execute it
		testcontainers.WithCmd("-c", "while [ ! -f "+starterScript+" ]; do sleep 0.1; done; bash "+starterScript),
		testcontainers.WithLifecycleHooks(testcontainers.ContainerLifecycleHooks{
			PostStarts: []testcontainers.ContainerHook{
				// Use a single hook to copy the starter script and wait for
				// the Kafka server to be ready. This prevents the wait running
				// if the starter script fails to copy.
				func(ctx context.Context, c testcontainers.Container) error {
					// 1. copy the SSL material and the starter script into the container
					if err := copyStarterScript(ctx, c, settings); err != nil {
						return fmt.Errorf("copy starter script: %w", err)
					}

					// 2. wait for the Kafka server to be ready
					return wait.ForLog(".*Transitioning from RECOVERY to RUNNING.*").AsRegexp().WaitUntilReady(ctx, c)
				},
			},
		}),
	)

	if settings.tlsEnabled {
		moduleOpts = append(moduleOpts,
			testcontainers.WithExposedPorts(sslPort),
			// the Confluent image reads the files below from /etc/kafka/secrets
			testcontainers.WithEnv(map[string]string{
				"KAFKA_SSL_KEYSTORE_FILENAME":    keystoreFilename,
				"KAFKA_SSL_KEYSTORE_CREDENTIALS": credentialsFilename,
				"KAFKA_SSL_KEY_CREDENTIALS":      credentialsFilename,
				"KAFKA_SSL_KEYSTORE_TYPE":        "PKCS12",
			}),
		)
	}

	moduleOpts = append(moduleOpts, opts...)

	// configure the controller quorum voters after all the options have been applied
	moduleOpts = append(moduleOpts, configureControllerQuorumVoters())

	// add the SSL listener after all the options have been applied,
	// so it is kept even if the listeners are overridden
	if settings.tlsEnabled {
		moduleOpts = append(moduleOpts, configureSSLListener())
	}

	var c *KafkaContainer
	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	if ctr != nil {
		c = &KafkaContainer{Container: ctr, settings: settings}
	}
	if err != nil {
		return c, fmt.Errorf("run kafka: %w", err)
	}

	// Inspect the container to get the CLUSTER_ID environment variable
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect kafka: %w", err)
	}

	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "CLUSTER_ID="); ok {
			c.ClusterID = v
			break
		}
	}

	return c, nil
}

// copyStarterScript copies the starter script into the container.
// If TLS is enabled, it first generates the certificates for the host the
// container is reachable at and copies the keystore into the container,
// so they are in place before Kafka starts.
func copyStarterScript(ctx context.Context, c testcontainers.Container, settings *options) error {
	if err := wait.ForMappedPort(publicPort).
		WaitUntilReady(ctx, c); err != nil {
		return fmt.Errorf("wait for mapped port: %w", err)
	}

	endpoint, err := c.PortEndpoint(ctx, publicPort, "PLAINTEXT")
	if err != nil {
		return fmt.Errorf("port endpoint: %w", err)
	}

	if settings.tlsEnabled {
		sslEndpoint, err := copyTLSMaterial(ctx, c, settings)
		if err != nil {
			return fmt.Errorf("copy TLS material: %w", err)
		}

		endpoint += "," + sslEndpoint
	}

	inspect, err := c.Inspect(ctx)
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}

	hostname := inspect.Config.Hostname

	scriptContent := fmt.Sprintf(starterScriptContent, endpoint, hostname)

	if err := c.CopyToContainer(ctx, []byte(scriptContent), starterScript, 0o755); err != nil {
		return fmt.Errorf("copy to container: %w", err)
	}

	return nil
}

// copyTLSMaterial generates the certificates, copies the keystore and its credentials
// into the container and returns the advertised SSL listener.
func copyTLSMaterial(ctx context.Context, c testcontainers.Container, settings *options) (string, error) {
	if err := wait.ForMappedPort(sslPort).WaitUntilReady(ctx, c); err != nil {
		return "", fmt.Errorf("wait for mapped SSL port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	certs, err := settings.ensureTLSCerts(host)
	if err != nil {
		return "", err
	}

	if err := c.CopyToContainer(ctx, certs.KeystoreBytes, secretsDir+"/"+keystoreFilename, 0o644); err != nil {
		return "", fmt.Errorf("copy keystore: %w", err)
	}

	if err := c.CopyToContainer(ctx, []byte(keystorePassword), secretsDir+"/"+credentialsFilename, 0o644); err != nil {
		return "", fmt.Errorf("copy keystore credentials: %w", err)
	}

	return c.PortEndpoint(ctx, sslPort, "SSL")
}

// WithClusterID sets the CLUSTER_ID environment variable, used to format the KRaft storage.
func WithClusterID(clusterID string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"CLUSTER_ID": clusterID,
	})
}

// Brokers retrieves the broker connection strings from Kafka with only one entry,
// defined by the exposed public port.
func (kc *KafkaContainer) Brokers(ctx context.Context) ([]string, error) {
	endpoint, err := kc.PortEndpoint(ctx, publicPort, "")
	if err != nil {
		return nil, err
	}

	return []string{endpoint}, nil
}

// BrokersTLS retrieves the broker connection strings for the SSL listener,
// defined by the exposed SSL port. Returns an error if TLS is not enabled.
func (kc *KafkaContainer) BrokersTLS(ctx context.Context) ([]string, error) {
	if kc.settings == nil || !kc.settings.tlsEnabled {
		return nil, errors.New("TLS is not enabled on this container")
	}

	endpoint, err := kc.PortEndpoint(ctx, sslPort, "")
	if err != nil {
		return nil, err
	}

	return []string{endpoint}, nil
}

// TLSConfig returns the TLS configuration for secure client connections,
// trusting the CA that signed the broker certificate.
// Returns an error if TLS is not enabled on this container.
func (kc *KafkaContainer) TLSConfig() (*tls.Config, error) {
	if kc.settings == nil || !kc.settings.tlsEnabled || kc.settings.tlsCerts == nil {
		return nil, errors.New("TLS is not enabled on this container")
	}

	return kc.settings.tlsCerts.TLSConfig.Clone(), nil
}

// configureControllerQuorumVoters returns an option that sets the quorum voters for the controller.
// For that, it will check if there are any network aliases defined for the container and use the
// first alias in the first network. Else, it will use localhost.
func configureControllerQuorumVoters() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
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

			req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"] = "1@" + host + ":9094"
		}

		return nil
	}
	// }
}

// configureSSLListener returns an option that adds the SSL listener to the listeners
// and to the listener security protocol map. If they already define it, it must
// listen on the SSL port and use the SSL protocol, as the module exposes and
// advertises that port. Otherwise, an error is returned.
func configureSSLListener() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = map[string]string{}
		}

		listeners := req.Env["KAFKA_LISTENERS"]
		if listener, ok := findListEntry(listeners, "SSL://"); ok {
			_, port, err := net.SplitHostPort(strings.TrimPrefix(listener, "SSL://"))
			if err != nil || port != sslPortNumber {
				return fmt.Errorf("KAFKA_LISTENERS defines the SSL listener %q, but WithTLS requires it to listen on port %s", listener, sslPortNumber)
			}
		} else {
			req.Env["KAFKA_LISTENERS"] = appendListEntry(listeners, "SSL://0.0.0.0:"+sslPortNumber)
		}

		protocolMap := req.Env["KAFKA_LISTENER_SECURITY_PROTOCOL_MAP"]
		if protocol, ok := findListEntry(protocolMap, "SSL:"); ok {
			if protocol != "SSL:SSL" {
				return fmt.Errorf("KAFKA_LISTENER_SECURITY_PROTOCOL_MAP maps the SSL listener as %q, but WithTLS requires %q", protocol, "SSL:SSL")
			}
		} else {
			req.Env["KAFKA_LISTENER_SECURITY_PROTOCOL_MAP"] = appendListEntry(protocolMap, "SSL:SSL")
		}

		return nil
	}
}

// findListEntry returns the first entry of the comma-separated list with the given prefix.
func findListEntry(list, prefix string) (string, bool) {
	for item := range strings.SplitSeq(list, ",") {
		if item = strings.TrimSpace(item); strings.HasPrefix(item, prefix) {
			return item, true
		}
	}

	return "", false
}

// appendListEntry appends the entry to the comma-separated list.
func appendListEntry(list, entry string) string {
	if list == "" {
		return entry
	}

	return list + "," + entry
}

// validateKRaftVersion validates if the image version is compatible with KRaft mode,
// which is available since version 7.0.0.
func validateKRaftVersion(fqName string) error {
	if fqName == "" {
		return errors.New("image cannot be empty")
	}

	idx := strings.LastIndex(fqName, ":")
	if idx == -1 || idx == len(fqName)-1 {
		return nil
	}

	image := fqName[:idx]
	version := fqName[idx+1:]

	if !strings.EqualFold(image, "confluentinc/confluent-local") {
		// do not validate if the image is not the official one.
		// not raising an error here, letting the image start and
		// eventually evaluate an error if it exists.
		return nil
	}

	// semver requires the version to start with a "v"
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	// remove the architecture suffix
	if strings.HasSuffix(version, ".amd64") || strings.HasSuffix(version, ".arm64") {
		return fmt.Errorf("invalid image tag %q: architecture suffixes like .arm64 or .amd64 are not valid semver; please use a multi-architecture image instead", version)
	}

	if semver.Compare(version, "v7.4.0") < 0 { // version < v7.4.0
		return fmt.Errorf("version=%s. KRaft mode is only available since version 7.4.0", version)
	}

	return nil
}
