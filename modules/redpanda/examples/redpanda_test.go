package redpanda_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mdelapenya/tlscert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/redpanda"
)

const testImage = "docker.redpanda.com/redpandadata/redpanda:v25.2.4"

func TestRedpanda(t *testing.T) {
	ctx := context.Background()

	ctr, err := redpanda.Run(ctx, testImage)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Test Kafka API
	seedBroker, err := ctr.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	kafkaCl, err := kgo.NewClient(
		kgo.SeedBrokers(seedBroker),
	)
	require.NoError(t, err)
	defer kafkaCl.Close()

	kafkaAdmCl := kadm.NewClient(kafkaCl)
	metadata, err := kafkaAdmCl.Metadata(ctx)
	require.NoError(t, err)
	require.Len(t, metadata.Brokers, 1)

	// Test Schema Registry API
	httpCl := &http.Client{Timeout: 5 * time.Second}
	schemaRegistryURL, err := ctr.SchemaRegistryAddress(ctx)
	require.NoError(t, err)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, schemaRegistryURL+"/subjects", nil)
	require.NoError(t, err)
	resp, err := httpCl.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Test Admin API
	// adminAPIAddress {
	adminAPIURL, err := ctr.AdminAPIAddress(ctx)
	// }
	require.NoError(t, err)
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, adminAPIURL+"/v1/cluster/health_overview", nil)
	require.NoError(t, err)
	resp, err = httpCl.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Test HTTP Proxy API
	// httpProxyAddress {
	httpProxyURL, err := ctr.HTTPProxyAddress(ctx)
	// }
	require.NoError(t, err)
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, httpProxyURL+"/topics", nil)
	require.NoError(t, err)
	resp, err = httpCl.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Test produce to unknown topic
	results := kafkaCl.ProduceSync(ctx, &kgo.Record{Topic: "test", Value: []byte("test message")})
	require.Error(t, results.FirstErr(), kerr.UnknownTopicOrPartition)
}

func TestRedpandaWithAuthentication(t *testing.T) {
	ctx := context.Background()
	// redpandaCreateContainer {
	ctr, err := redpanda.Run(ctx,
		testImage,
		redpanda.WithEnableSASL(),
		redpanda.WithEnableKafkaAuthorization(),
		redpanda.WithEnableWasmTransform(),
		redpanda.WithNewServiceAccount("superuser-1", "test"),
		redpanda.WithNewServiceAccount("superuser-2", "test"),
		redpanda.WithNewServiceAccount("no-superuser", "test"),
		redpanda.WithSuperusers("superuser-1", "superuser-2"),
		redpanda.WithEnableSchemaRegistryHTTPBasicAuth(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	// }

	// kafkaSeedBroker {
	seedBroker, err := ctr.KafkaSeedBroker(ctx)
	// }
	require.NoError(t, err)

	// Schema Registry Client
	httpCl := &http.Client{Timeout: 5 * time.Second}
	// schemaRegistryAddress {
	schemaRegistryURL, err := ctr.SchemaRegistryAddress(ctx)
	// }
	require.NoError(t, err)

	serviceAccounts := map[string]string{
		"superuser-1": "test",
		"superuser-2": "test",
	}

	// Test successful authentication & authorization with all created superusers
	t.Run("happy-path", func(t *testing.T) {
		for user, password := range serviceAccounts {
			kafkaCl, err := kgo.NewClient(
				kgo.SeedBrokers(seedBroker),
				kgo.SASL(scram.Auth{
					User: user,
					Pass: password,
				}.AsSha256Mechanism()),
			)
			require.NoError(t, err)

			kafkaAdmCl := kadm.NewClient(kafkaCl)
			_, err = kafkaAdmCl.CreateTopic(ctx, 1, 1, nil, fmt.Sprintf("test-%v", user))
			kafkaCl.Close()
			require.NoError(t, err)
		}
	})

	// Test successful authentication, but failed authorization with a non-superuser account
	t.Run("non-superuser account", func(t *testing.T) {
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(scram.Auth{
				User: "no-superuser",
				Pass: "test",
			}.AsSha256Mechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.CreateTopic(ctx, 1, 1, nil, "test-2")
		kafkaCl.Close()
		require.Error(t, err)
		require.ErrorContains(t, err, "TOPIC_AUTHORIZATION_FAILED")
	})

	// Test failed authentication
	t.Run("invalid-user", func(t *testing.T) {
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(scram.Auth{
				User: "wrong",
				Pass: "wrong",
			}.AsSha256Mechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.Metadata(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "SASL_AUTHENTICATION_FAILED")
	})

	// Test Schema Registry API
	t.Run("schema-registry", func(t *testing.T) {
		t.Run("failed-authentication", func(t *testing.T) {
			// Failed authentication
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, schemaRegistryURL+"/subjects", nil)
			require.NoError(t, err)
			resp, err := httpCl.Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			resp.Body.Close()
		})

		t.Run("successful-authentication", func(t *testing.T) {
			for user, password := range serviceAccounts {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, schemaRegistryURL+"/subjects", nil)
				require.NoError(t, err)
				req.SetBasicAuth(user, password)
				resp, err := httpCl.Do(req)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		})
	})
}

func TestRedpandaWithBootstrapUserAuthentication(t *testing.T) {
	ctx := context.Background()
	ctr, err := redpanda.Run(ctx,
		testImage,
		redpanda.WithEnableSASL(),
		redpanda.WithEnableKafkaAuthorization(),
		redpanda.WithEnableWasmTransform(),
		redpanda.WithNewServiceAccount("superuser-1", "test"),
		redpanda.WithNewServiceAccount("superuser-2", "test"),
		redpanda.WithNewServiceAccount("no-superuser", "test"),
		redpanda.WithSuperusers("superuser-1", "superuser-2"),
		redpanda.WithEnableSchemaRegistryHTTPBasicAuth(),
		redpanda.WithAdminAPIAuthentication(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	seedBroker, err := ctr.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	serviceAccounts := map[string]string{
		"superuser-1": "test",
		"superuser-2": "test",
	}

	// Schema Registry Client
	httpCl := &http.Client{Timeout: 5 * time.Second}
	schemaRegistryURL, err := ctr.SchemaRegistryAddress(ctx)
	require.NoError(t, err)

	// Test successful authentication & authorization with all created superusers
	t.Run("happy-path", func(t *testing.T) {
		for user, password := range serviceAccounts {
			kafkaCl, err := kgo.NewClient(
				kgo.SeedBrokers(seedBroker),
				kgo.SASL(scram.Auth{
					User: user,
					Pass: password,
				}.AsSha256Mechanism()),
			)
			require.NoError(t, err)

			kafkaAdmCl := kadm.NewClient(kafkaCl)
			_, err = kafkaAdmCl.CreateTopic(ctx, 1, 1, nil, fmt.Sprintf("test-%v", user))
			kafkaCl.Close()
			require.NoError(t, err)
		}
	})

	// Test successful authentication, but failed authorization with a non-superuser account
	t.Run("non-superuser", func(t *testing.T) {
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(scram.Auth{
				User: "no-superuser",
				Pass: "test",
			}.AsSha256Mechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.CreateTopic(ctx, 1, 1, nil, "test-2")
		kafkaCl.Close()
		require.Error(t, err)
		require.ErrorContains(t, err, "TOPIC_AUTHORIZATION_FAILED")
	})

	// Test failed authentication
	t.Run("invalid-user", func(t *testing.T) {
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(scram.Auth{
				User: "wrong",
				Pass: "wrong",
			}.AsSha256Mechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.Metadata(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "SASL_AUTHENTICATION_FAILED")
	})

	// Test Schema Registry API
	t.Run("schema-registry", func(t *testing.T) {
		t.Run("failed authentication", func(t *testing.T) {
			// Failed authentication
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, schemaRegistryURL+"/subjects", nil)
			require.NoError(t, err)
			resp, err := httpCl.Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			resp.Body.Close()
		})

		t.Run("successful-authentication", func(t *testing.T) {
			for user, password := range serviceAccounts {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, schemaRegistryURL+"/subjects", nil)
				require.NoError(t, err)
				req.SetBasicAuth(user, password)
				resp, err := httpCl.Do(req)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		})
	})
}

func TestRedpandaWithOldVersionAndWasm(t *testing.T) {
	ctx := context.Background()
	// this would fail to start if we weren't ignoring wasm transforms for older versions
	ctr, err := redpanda.Run(ctx,
		"redpandadata/redpanda:v23.2.18",
		redpanda.WithEnableSASL(),
		redpanda.WithEnableKafkaAuthorization(),
		redpanda.WithEnableWasmTransform(),
		redpanda.WithNewServiceAccount("superuser-1", "test"),
		redpanda.WithNewServiceAccount("superuser-2", "test"),
		redpanda.WithNewServiceAccount("no-superuser", "test"),
		redpanda.WithSuperusers("superuser-1", "superuser-2"),
		redpanda.WithEnableSchemaRegistryHTTPBasicAuth(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	seedBroker, err := ctr.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	// Schema Registry client
	httpCl := &http.Client{Timeout: 5 * time.Second}
	schemaRegistryURL, err := ctr.SchemaRegistryAddress(ctx)
	require.NoError(t, err)

	// Test successful authentication & authorization with all created superusers
	serviceAccounts := map[string]string{
		"superuser-1": "test",
		"superuser-2": "test",
	}

	for user, password := range serviceAccounts {
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(scram.Auth{
				User: user,
				Pass: password,
			}.AsSha256Mechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.CreateTopic(ctx, 1, 1, nil, fmt.Sprintf("test-%v", user))
		kafkaCl.Close()
		require.NoError(t, err)
	}

	// Test successful authentication, but failed authorization with a non-superuser account
	t.Run("non-superuser", func(t *testing.T) {
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(scram.Auth{
				User: "no-superuser",
				Pass: "test",
			}.AsSha256Mechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.CreateTopic(ctx, 1, 1, nil, "test-2")
		kafkaCl.Close()
		require.Error(t, err)
		require.ErrorContains(t, err, "TOPIC_AUTHORIZATION_FAILED")
	})

	// Test failed authentication
	t.Run("invalid-user", func(t *testing.T) {
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(scram.Auth{
				User: "wrong",
				Pass: "wrong",
			}.AsSha256Mechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.Metadata(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "SASL_AUTHENTICATION_FAILED")
	})

	// Test wrong mechanism
	t.Run("wrong-mechanism", func(t *testing.T) {
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(plain.Auth{
				User: "no-superuser",
				Pass: "test",
			}.AsMechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.Metadata(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "UNSUPPORTED_SASL_MECHANISM")
	})

	// Test Schema Registry API
	t.Run("schema-registry", func(t *testing.T) {
		t.Run("failed-authentication", func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, schemaRegistryURL+"/subjects", nil)
			require.NoError(t, err)
			resp, err := httpCl.Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			resp.Body.Close()
		})

		// Successful authentication
		t.Run("successful-authentication", func(t *testing.T) {
			for user, password := range serviceAccounts {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, schemaRegistryURL+"/subjects", nil)
				require.NoError(t, err)
				req.SetBasicAuth(user, password)
				resp, err := httpCl.Do(req)
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		})
	})
}

func TestRedpandaProduceWithAutoCreateTopics(t *testing.T) {
	ctx := context.Background()

	ctr, err := redpanda.Run(ctx, testImage, redpanda.WithAutoCreateTopics())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	brokers, err := ctr.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	kafkaCl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers),
		kgo.AllowAutoTopicCreation(),
	)
	require.NoError(t, err)
	defer kafkaCl.Close()

	results := kafkaCl.ProduceSync(ctx, &kgo.Record{Topic: "test", Value: []byte("test message")})
	require.NoError(t, results.FirstErr())
}

func TestRedpandaWithTLS(t *testing.T) {
	ctx := context.Background()

	containerHostAddress, err := containerHost(ctx)
	require.NoError(t, err)
	cert, err := tlscert.SelfSignedFromRequestE(tlscert.Request{
		Name: "client",
		Host: "localhost,127.0.0.1," + containerHostAddress,
	})
	require.NoError(t, err, "failed to generate cert")

	ctr, err := redpanda.Run(ctx, testImage, redpanda.WithTLS(cert.Bytes, cert.KeyBytes))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	tlsConfig := cert.TLSConfig()

	httpCl := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			ForceAttemptHTTP2:   true,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     tlsConfig,
		},
	}

	// Test Admin API
	adminAPIURL, err := ctr.AdminAPIAddress(ctx)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(adminAPIURL, "https://"), "AdminAPIAddress should return https url")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, adminAPIURL+"/v1/cluster/health_overview", nil)
	require.NoError(t, err)
	resp, err := httpCl.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test Schema Registry API
	schemaRegistryURL, err := ctr.SchemaRegistryAddress(ctx)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(adminAPIURL, "https://"), "SchemaRegistryAddress should return https url")
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, schemaRegistryURL+"/subjects", nil)
	require.NoError(t, err)
	resp, err = httpCl.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	brokers, err := ctr.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	kafkaCl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers),
		kgo.DialTLSConfig(tlsConfig),
	)
	require.NoError(t, err)
	defer kafkaCl.Close()

	// Test produce to unknown topic
	results := kafkaCl.ProduceSync(ctx, &kgo.Record{Topic: "test", Value: []byte("test message")})
	require.Error(t, results.FirstErr(), kerr.UnknownTopicOrPartition)
}

func TestRedpandaWithTLSAndSASL(t *testing.T) {
	ctx := context.Background()

	containerHostAddress, err := containerHost(ctx)
	require.NoError(t, err)
	cert, err := tlscert.SelfSignedFromRequestE(tlscert.Request{
		Name: "client",
		Host: "localhost,127.0.0.1," + containerHostAddress,
	})
	require.NoError(t, err, "failed to generate cert")

	ctr, err := redpanda.Run(ctx,
		testImage,
		redpanda.WithTLS(cert.Bytes, cert.KeyBytes),
		redpanda.WithEnableSASL(),
		redpanda.WithEnableKafkaAuthorization(),
		redpanda.WithNewServiceAccount("superuser-1", "test"),
		redpanda.WithSuperusers("superuser-1"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	tlsConfig := cert.TLSConfig()

	broker, err := ctr.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	kafkaCl, err := kgo.NewClient(
		kgo.SeedBrokers(broker),
		kgo.DialTLSConfig(tlsConfig),
		kgo.SASL(scram.Auth{
			User: "superuser-1",
			Pass: "test",
		}.AsSha256Mechanism()),
	)
	require.NoError(t, err)
	defer kafkaCl.Close()

	_, err = kadm.NewClient(kafkaCl).ListTopics(ctx)
	require.NoError(t, err)
}

func containerHost(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (string, error) {
	// Use a dummy request to get the provider from options.
	var req testcontainers.GenericContainerRequest
	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return "", err
		}
	}

	logging := req.Logger
	if logging == nil {
		logging = log.Default()
	}
	p, err := req.ProviderType.GetProvider(testcontainers.WithLogger(logging))
	if err != nil {
		return "", err
	}

	defer p.Close()

	if p, ok := p.(*testcontainers.DockerProvider); ok {
		return p.DaemonHost(ctx)
	}

	// Fall back to localhost.
	return "localhost", nil
}
