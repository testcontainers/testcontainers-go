package redpanda_test

import (
	"context"
	"fmt"
	"io"
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
	"github.com/testcontainers/testcontainers-go/modules/redpanda"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestRedpanda(t *testing.T) {
	ctx := context.Background()

	container, err := redpanda.RunContainer(ctx)
	require.NoError(t, err)

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// Test Kafka API
	seedBroker, err := container.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	kafkaCl, err := kgo.NewClient(
		kgo.SeedBrokers(seedBroker),
	)
	require.NoError(t, err)
	defer kafkaCl.Close()

	kafkaAdmCl := kadm.NewClient(kafkaCl)
	metadata, err := kafkaAdmCl.Metadata(ctx)
	require.NoError(t, err)
	assert.Len(t, metadata.Brokers, 1)

	// Test Schema Registry API
	httpCl := &http.Client{Timeout: 5 * time.Second}
	schemaRegistryURL, err := container.SchemaRegistryAddress(ctx)
	require.NoError(t, err)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/subjects", schemaRegistryURL), nil)
	require.NoError(t, err)
	resp, err := httpCl.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test Admin API
	// adminAPIAddress {
	adminAPIURL, err := container.AdminAPIAddress(ctx)
	// }
	require.NoError(t, err)
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/cluster/health_overview", adminAPIURL), nil)
	require.NoError(t, err)
	resp, err = httpCl.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test produce to unknown topic
	results := kafkaCl.ProduceSync(ctx, &kgo.Record{Topic: "test", Value: []byte("test message")})
	require.Error(t, results.FirstErr(), kerr.UnknownTopicOrPartition)
}

func TestRedpandaWithAuthentication(t *testing.T) {
	ctx := context.Background()
	// redpandaCreateContainer {
	container, err := redpanda.RunContainer(ctx,
		redpanda.WithEnableSASL(),
		redpanda.WithEnableKafkaAuthorization(),
		redpanda.WithEnableWasmTransform(),
		redpanda.WithNewServiceAccount("superuser-1", "test"),
		redpanda.WithNewServiceAccount("superuser-2", "test"),
		redpanda.WithNewServiceAccount("no-superuser", "test"),
		redpanda.WithSuperusers("superuser-1", "superuser-2"),
		redpanda.WithEnableSchemaRegistryHTTPBasicAuth(),
	)
	require.NoError(t, err)
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// kafkaSeedBroker {
	seedBroker, err := container.KafkaSeedBroker(ctx)
	// }
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
		require.NoError(t, err)
		kafkaCl.Close()
	}

	// Test successful authentication, but failed authorization with a non-superuser account
	{
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
		require.Error(t, err)
		require.ErrorContains(t, err, "TOPIC_AUTHORIZATION_FAILED")
		kafkaCl.Close()
	}

	// Test failed authentication
	{
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
	}

	// Test Schema Registry API
	httpCl := &http.Client{Timeout: 5 * time.Second}
	// schemaRegistryAddress {
	schemaRegistryURL, err := container.SchemaRegistryAddress(ctx)
	// }
	require.NoError(t, err)

	// Failed authentication
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/subjects", schemaRegistryURL), nil)
	require.NoError(t, err)
	resp, err := httpCl.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Successful authentication
	for user, password := range serviceAccounts {
		req.SetBasicAuth(user, password)
		resp, err = httpCl.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestRedpandaWithOldVersionAndWasm(t *testing.T) {
	ctx := context.Background()
	// redpandaCreateContainer {
	// this would fail to start if we weren't ignoring wasm transforms for older versions
	container, err := redpanda.RunContainer(ctx,
		testcontainers.WithImage("redpandadata/redpanda:v23.2.18"),
		redpanda.WithEnableSASL(),
		redpanda.WithEnableKafkaAuthorization(),
		redpanda.WithEnableWasmTransform(),
		redpanda.WithNewServiceAccount("superuser-1", "test"),
		redpanda.WithNewServiceAccount("superuser-2", "test"),
		redpanda.WithNewServiceAccount("no-superuser", "test"),
		redpanda.WithSuperusers("superuser-1", "superuser-2"),
		redpanda.WithEnableSchemaRegistryHTTPBasicAuth(),
	)
	require.NoError(t, err)
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// kafkaSeedBroker {
	seedBroker, err := container.KafkaSeedBroker(ctx)
	// }
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
		require.NoError(t, err)
		kafkaCl.Close()
	}

	// Test successful authentication, but failed authorization with a non-superuser account
	{
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
		require.Error(t, err)
		require.ErrorContains(t, err, "TOPIC_AUTHORIZATION_FAILED")
		kafkaCl.Close()
	}

	// Test failed authentication
	{
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
	}

	// Test wrong mechanism
	{
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
	}

	// Test Schema Registry API
	httpCl := &http.Client{Timeout: 5 * time.Second}
	// schemaRegistryAddress {
	schemaRegistryURL, err := container.SchemaRegistryAddress(ctx)
	// }
	require.NoError(t, err)

	// Failed authentication
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/subjects", schemaRegistryURL), nil)
	require.NoError(t, err)
	resp, err := httpCl.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Successful authentication
	for user, password := range serviceAccounts {
		req.SetBasicAuth(user, password)
		resp, err = httpCl.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestRedpandaProduceWithAutoCreateTopics(t *testing.T) {
	ctx := context.Background()

	container, err := redpanda.RunContainer(ctx, redpanda.WithAutoCreateTopics())
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	brokers, err := container.KafkaSeedBroker(ctx)
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
	tmp := t.TempDir()
	cert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:      "client",
		Host:      "localhost,127.0.0.1",
		ParentDir: tmp,
	})
	require.NotNil(t, cert, "failed to generate cert")

	ctx := context.Background()

	container, err := redpanda.RunContainer(ctx, redpanda.WithTLS(cert.Bytes, cert.KeyBytes))
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

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
	adminAPIURL, err := container.AdminAPIAddress(ctx)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(adminAPIURL, "https://"), "AdminAPIAddress should return https url")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/cluster/health_overview", adminAPIURL), nil)
	require.NoError(t, err)
	resp, err := httpCl.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test Schema Registry API
	schemaRegistryURL, err := container.SchemaRegistryAddress(ctx)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(adminAPIURL, "https://"), "SchemaRegistryAddress should return https url")
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/subjects", schemaRegistryURL), nil)
	require.NoError(t, err)
	resp, err = httpCl.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	brokers, err := container.KafkaSeedBroker(ctx)
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
	tmp := t.TempDir()

	cert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:      "client",
		Host:      "localhost,127.0.0.1",
		ParentDir: tmp,
	})
	require.NotNil(t, cert, "failed to generate cert")

	ctx := context.Background()

	container, err := redpanda.RunContainer(ctx,
		redpanda.WithTLS(cert.Bytes, cert.KeyBytes),
		redpanda.WithEnableSASL(),
		redpanda.WithEnableKafkaAuthorization(),
		redpanda.WithNewServiceAccount("superuser-1", "test"),
		redpanda.WithSuperusers("superuser-1"),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	tlsConfig := cert.TLSConfig()

	broker, err := container.KafkaSeedBroker(ctx)
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

func TestRedpandaListener_Simple(t *testing.T) {
	ctx := context.Background()

	// 1. Create network
	rpNetwork, err := network.New(ctx, network.WithCheckDuplicate())
	require.NoError(t, err)

	// 2. Start Redpanda container
	// withListenerRP {
	container, err := redpanda.RunContainer(ctx,
		testcontainers.WithImage("redpandadata/redpanda:v23.2.18"),
		network.WithNetwork([]string{"redpanda-host"}, rpNetwork),
		redpanda.WithListener("redpanda:29092"), redpanda.WithAutoCreateTopics(),
	)
	// }
	require.NoError(t, err)

	// 3. Start KCat container
	// withListenerKcat {
	kcat, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "confluentinc/cp-kcat:7.4.1",
			Networks: []string{
				rpNetwork.Name,
			},
			Entrypoint: []string{
				"sh",
			},
			Cmd: []string{
				"-c",
				"tail -f /dev/null",
			},
		},
		Started: true,
	})
	// }

	require.NoError(t, err)

	// 4. Copy message to kcat
	err = kcat.CopyToContainer(ctx, []byte("Message produced by kcat"), "/tmp/msgs.txt", 700)
	require.NoError(t, err)

	// 5. Produce message to Redpanda
	// withListenerExec {
	_, _, err = kcat.Exec(ctx, []string{"kcat", "-b", "redpanda:29092", "-t", "msgs", "-P", "-l", "/tmp/msgs.txt"})
	// }

	require.NoError(t, err)

	// 6. Consume message from Redpanda
	_, stdout, err := kcat.Exec(ctx, []string{"kcat", "-b", "redpanda:29092", "-C", "-t", "msgs", "-c", "1"})
	require.NoError(t, err)

	// 7. Read Message from stdout
	out, err := io.ReadAll(stdout)
	require.NoError(t, err)

	require.Contains(t, string(out), "Message produced by kcat")

	t.Cleanup(func() {
		if err := kcat.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate kcat container: %s", err)
		}
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate redpanda container: %s", err)
		}

		if err := rpNetwork.Remove(ctx); err != nil {
			t.Fatalf("failed to remove network: %s", err)
		}
	})
}

func TestRedpandaListener_InvalidPort(t *testing.T) {
	ctx := context.Background()

	// 1. Create network
	RPNetwork, err := network.New(ctx, network.WithCheckDuplicate())
	require.NoError(t, err)

	// 2. Attempt Start Redpanda container
	_, err = redpanda.RunContainer(ctx,
		testcontainers.WithImage("redpandadata/redpanda:v23.2.18"),
		redpanda.WithListener("redpanda:99092"),
		network.WithNetwork([]string{"redpanda-host"}, RPNetwork),
	)

	require.Error(t, err)

	require.Contains(t, err.Error(), "invalid port on listener redpanda:99092")

	t.Cleanup(func() {
		if err := RPNetwork.Remove(ctx); err != nil {
			t.Fatalf("failed to remove network: %s", err)
		}
	})
}

func TestRedpandaListener_NoNetwork(t *testing.T) {
	ctx := context.Background()

	// 1. Attempt Start Redpanda container
	_, err := redpanda.RunContainer(ctx,
		testcontainers.WithImage("redpandadata/redpanda:v23.2.18"),
		redpanda.WithListener("redpanda:99092"),
	)

	require.Error(t, err)

	require.Contains(t, err.Error(), "container must be attached to at least one network")
}
