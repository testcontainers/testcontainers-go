package rabbitmq_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mdelapenya/tlscert"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func TestRunContainer_connectUsingAmqp(t *testing.T) {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.Run(ctx, "rabbitmq:3.12.11-management-alpine")
	testcontainers.CleanupContainer(t, rabbitmqContainer)
	require.NoError(t, err)

	amqpURL, err := rabbitmqContainer.AmqpURL(ctx)
	require.NoError(t, err)

	amqpConnection, err := amqp.Dial(amqpURL)
	require.NoError(t, err)

	err = amqpConnection.Close()
	require.NoError(t, err)
}

func TestRunContainer_connectUsingAmqps(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()

	caCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:      "ca",
		Host:      "localhost,127.0.0.1",
		IsCA:      true,
		ParentDir: tmpDir,
	})
	require.NotNilf(t, caCert, "failed to generate CA certificate")

	cert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:      "client",
		Host:      "localhost,127.0.0.1",
		IsCA:      true,
		Parent:    caCert,
		ParentDir: tmpDir,
	})
	require.NotNilf(t, cert, "failed to generate certificate")

	sslSettings := rabbitmq.SSLSettings{
		CACertFile:        caCert.CertPath,
		CertFile:          cert.CertPath,
		KeyFile:           cert.KeyPath,
		VerificationMode:  rabbitmq.SSLVerificationModePeer,
		FailIfNoCert:      false,
		VerificationDepth: 1,
	}

	rabbitmqContainer, err := rabbitmq.Run(ctx, "rabbitmq:3.12.11-management-alpine", rabbitmq.WithSSL(sslSettings))
	testcontainers.CleanupContainer(t, rabbitmqContainer)
	require.NoError(t, err)

	amqpsURL, err := rabbitmqContainer.AmqpsURL(ctx)
	require.NoError(t, err)

	require.Truef(t, strings.HasPrefix(amqpsURL, "amqps"), "AMQPS Url should begin with `amqps`")

	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(sslSettings.CACertFile)
	require.NoError(t, err)
	certs.AppendCertsFromPEM(pemData)

	amqpsConnection, err := amqp.DialTLS(amqpsURL, &tls.Config{InsecureSkipVerify: false, RootCAs: certs})
	require.NoError(t, err)

	require.Falsef(t, amqpsConnection.IsClosed(), "AMQPS Connection unexpectdely closed")
	err = amqpsConnection.Close()
	require.NoError(t, err)
}

func TestRunContainer_withAllSettings(t *testing.T) {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:3.12.11-management-alpine",
		// addVirtualHosts {
		testcontainers.WithAfterReadyCommand(VirtualHost{Name: "vhost1"}),
		testcontainers.WithAfterReadyCommand(VirtualHostLimit{VHost: "vhost1", Name: "max-connections", Value: 1}),
		testcontainers.WithAfterReadyCommand(VirtualHost{Name: "vhost2", Tracing: true}),
		// }
		// addExchanges {
		testcontainers.WithAfterReadyCommand(Exchange{Name: "direct-exchange", Type: "direct"}),
		testcontainers.WithAfterReadyCommand(Exchange{
			Name: "topic-exchange",
			Type: "topic",
		}),
		testcontainers.WithAfterReadyCommand(Exchange{
			VHost:      "vhost1",
			Name:       "topic-exchange-2",
			Type:       "topic",
			AutoDelete: false,
			Internal:   false,
			Durable:    true,
			Args:       map[string]interface{}{},
		}),
		testcontainers.WithAfterReadyCommand(Exchange{
			VHost: "vhost2",
			Name:  "topic-exchange-3",
			Type:  "topic",
		}),
		testcontainers.WithAfterReadyCommand(Exchange{
			Name:       "topic-exchange-4",
			Type:       "topic",
			AutoDelete: false,
			Internal:   false,
			Durable:    true,
			Args:       map[string]interface{}{},
		}),
		// }
		// addQueues {
		testcontainers.WithAfterReadyCommand(Queue{Name: "queue1"}),
		testcontainers.WithAfterReadyCommand(Queue{
			Name:       "queue2",
			AutoDelete: true,
			Durable:    false,
			Args:       map[string]interface{}{"x-message-ttl": 1000},
		}),
		testcontainers.WithAfterReadyCommand(Queue{
			VHost:      "vhost1",
			Name:       "queue3",
			AutoDelete: true,
			Durable:    false,
			Args:       map[string]interface{}{"x-message-ttl": 1000},
		}),
		testcontainers.WithAfterReadyCommand(Queue{VHost: "vhost2", Name: "queue4"}),
		// }
		// addBindings {
		testcontainers.WithAfterReadyCommand(NewBinding("direct-exchange", "queue1")),
		testcontainers.WithAfterReadyCommand(NewBindingWithVHost("vhost1", "topic-exchange-2", "queue3")),
		testcontainers.WithAfterReadyCommand(Binding{
			VHost:           "vhost2",
			Source:          "topic-exchange-3",
			Destination:     "queue4",
			RoutingKey:      "ss7",
			DestinationType: "queue",
			Args:            map[string]interface{}{},
		}),
		// }
		// addUsers {
		testcontainers.WithAfterReadyCommand(User{
			Name:     "user1",
			Password: "password1",
		}),
		testcontainers.WithAfterReadyCommand(User{
			Name:     "user2",
			Password: "password2",
			Tags:     []string{"administrator"},
		}),
		// }
		// addPermissions {
		testcontainers.WithAfterReadyCommand(NewPermission("vhost1", "user1", ".*", ".*", ".*")),
		// }
		// addPolicies {
		testcontainers.WithAfterReadyCommand(Policy{
			Name:       "max length policy",
			Pattern:    "^dog",
			Definition: map[string]interface{}{"max-length": 1},
			Priority:   1,
			ApplyTo:    "queues",
		}),
		testcontainers.WithAfterReadyCommand(Policy{
			Name:       "alternate exchange policy",
			Pattern:    "^direct-exchange",
			Definition: map[string]interface{}{"alternate-exchange": "amq.direct"},
		}),
		testcontainers.WithAfterReadyCommand(Policy{
			VHost:   "vhost2",
			Name:    "ha-all",
			Pattern: ".*",
			Definition: map[string]interface{}{
				"ha-mode":      "all",
				"ha-sync-mode": "automatic",
			},
		}),
		testcontainers.WithAfterReadyCommand(OperatorPolicy{
			Name:       "operator policy 1",
			Pattern:    "^queue1",
			Definition: map[string]interface{}{"message-ttl": 1000},
			Priority:   1,
			ApplyTo:    "queues",
		}),
		// }
		// enablePlugins {
		testcontainers.WithAfterReadyCommand(Plugin{Name: "rabbitmq_shovel"}, Plugin{Name: "rabbitmq_random_exchange"}),
		// }
	)
	testcontainers.CleanupContainer(t, rabbitmqContainer)
	require.NoError(t, err)

	requireEntity(t, rabbitmqContainer, "queues", "queue1", "queue2", "queue3", "queue4")
	requireEntity(t, rabbitmqContainer, "exchanges", "direct-exchange", "topic-exchange", "topic-exchange-2", "topic-exchange-3", "topic-exchange-4")
	requireEntity(t, rabbitmqContainer, "users", "user1", "user2")
	requireEntity(t, rabbitmqContainer, "policies", "max length policy", "alternate exchange policy")
	requireEntityWithVHost(t, rabbitmqContainer, "policies", 2, "max length policy", "alternate exchange policy")
	requireEntity(t, rabbitmqContainer, "operator_policies", "operator policy 1")
	requirePluginIsEnabled(t, rabbitmqContainer, "rabbitmq_shovel", "rabbitmq_random_exchange")
}

func requireEntity(t *testing.T, container testcontainers.Container, listCommand string, entities ...string) {
	t.Helper()

	ctx := context.Background()

	cmd := []string{"rabbitmqadmin", "list", listCommand}

	check := testcontainers.RequireContainerExec(ctx, t, container, cmd)
	for _, e := range entities {
		require.Contains(t, check, e)
	}
}

func requireEntityWithVHost(t *testing.T, container testcontainers.Container, listCommand string, vhostID int, entities ...string) {
	t.Helper()

	ctx := context.Background()

	cmd := []string{"rabbitmqadmin", "list", listCommand}
	if vhostID > 0 {
		cmd = append(cmd, fmt.Sprintf("--vhost=vhost%d", vhostID))
	}

	check := testcontainers.RequireContainerExec(ctx, t, container, cmd)
	for _, e := range entities {
		require.Contains(t, check, e)
	}
}

func requirePluginIsEnabled(t *testing.T, container testcontainers.Container, plugins ...string) {
	t.Helper()

	ctx := context.Background()

	for _, plugin := range plugins {

		cmd := []string{"rabbitmq-plugins", "is_enabled", plugin}

		check := testcontainers.RequireContainerExec(ctx, t, container, cmd)
		require.Contains(t, check, plugin+" is enabled")
	}
}
