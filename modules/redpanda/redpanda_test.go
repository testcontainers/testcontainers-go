package redpanda

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

func TestRedpanda(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx)
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

	kafkaAdmCl := kadm.NewClient(kafkaCl)
	metadata, err := kafkaAdmCl.Metadata(ctx)
	require.NoError(t, err)
	assert.Len(t, metadata.Brokers, 1)

	// Test Schema Registry API
	httpCl := &http.Client{Timeout: 5 * time.Second}
	schemaRegistryURL, err := container.SchemaRegistryAddress(ctx)
	require.NoError(t, err)
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/subjects", schemaRegistryURL), nil)
	require.NoError(t, err)
	resp, err := httpCl.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test Admin API
	// adminAPIAddress {
	adminAPIURL, err := container.AdminAPIAddress(ctx)
	require.NoError(t, err)
	// }
	req, err = http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/cluster/health_overview", adminAPIURL), nil)
	require.NoError(t, err)
	resp, err = httpCl.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRedpandaWithAuthentication(t *testing.T) {
	ctx := context.Background()
	// redpandaCreateContainer {
	container, err := RunContainer(ctx,
		WithEnableSASL(),
		WithEnableKafkaAuthorization(),
		WithNewServiceAccount("superuser-1", "test"),
		WithNewServiceAccount("superuser-2", "test"),
		WithNewServiceAccount("no-superuser", "test"),
		WithSuperusers("superuser-1", "superuser-2"),
		WithEnableSchemaRegistryHTTPBasicAuth(),
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
	require.NoError(t, err)
	// }

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
		assert.ErrorContains(t, err, "TOPIC_AUTHORIZATION_FAILED")
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
		assert.ErrorContains(t, err, "SASL_AUTHENTICATION_FAILED")
	}

	// Test Schema Registry API
	httpCl := &http.Client{Timeout: 5 * time.Second}
	// schemaRegistryAddress {
	schemaRegistryURL, err := container.SchemaRegistryAddress(ctx)
	require.NoError(t, err)
	// }

	// Failed authentication
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/subjects", schemaRegistryURL), nil)
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
