package rabbitmq_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mdelapenya/tlscert"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func TestRunContainer_connectUsingAmqp(t *testing.T) {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	amqpURL, err := rabbitmqContainer.AmqpURL(ctx)
	if err != nil {
		t.Fatal(err)
	}

	amqpConnection, err := amqp.Dial(amqpURL)
	if err != nil {
		t.Fatal(err)
	}

	if err = amqpConnection.Close(); err != nil {
		t.Fatal(err)
	}
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
	if caCert == nil {
		t.Fatal("failed to generate CA certificate")
	}

	cert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:      "client",
		Host:      "localhost,127.0.0.1",
		IsCA:      true,
		Parent:    caCert,
		ParentDir: tmpDir,
	})
	if cert == nil {
		t.Fatal("failed to generate certificate")
	}

	sslSettings := rabbitmq.SSLSettings{
		CACertFile:        caCert.CertPath,
		CertFile:          cert.CertPath,
		KeyFile:           cert.KeyPath,
		VerificationMode:  rabbitmq.SSLVerificationModePeer,
		FailIfNoCert:      false,
		VerificationDepth: 1,
	}

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx, rabbitmq.WithSSL(sslSettings))
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	amqpsURL, err := rabbitmqContainer.AmqpsURL(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(amqpsURL, "amqps") {
		t.Fatal(fmt.Errorf("AMQPS Url should begin with `amqps`"))
	}

	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(sslSettings.CACertFile)
	if err != nil {
		t.Fatal(err)
	}
	certs.AppendCertsFromPEM(pemData)

	amqpsConnection, err := amqp.DialTLS(amqpsURL, &tls.Config{InsecureSkipVerify: false, RootCAs: certs})
	if err != nil {
		t.Fatal(err)
	}

	if amqpsConnection.IsClosed() {
		t.Fatal(fmt.Errorf("AMQPS Connection unexpectdely closed"))
	}
	if err = amqpsConnection.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRunContainer_withAllSettings(t *testing.T) {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.12.11-management-alpine"),
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
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	if !assertEntity(t, rabbitmqContainer, "queues", "queue1", "queue2", "queue3", "queue4") {
		t.Fatal(err)
	}
	if !assertEntity(t, rabbitmqContainer, "exchanges", "direct-exchange", "topic-exchange", "topic-exchange-2", "topic-exchange-3", "topic-exchange-4") {
		t.Fatal(err)
	}
	if !assertEntity(t, rabbitmqContainer, "users", "user1", "user2") {
		t.Fatal(err)
	}
	if !assertEntity(t, rabbitmqContainer, "policies", "max length policy", "alternate exchange policy") {
		t.Fatal(err)
	}
	if !assertEntityWithVHost(t, rabbitmqContainer, "policies", 2, "max length policy", "alternate exchange policy") {
		t.Fatal(err)
	}
	if !assertEntity(t, rabbitmqContainer, "operator_policies", "operator policy 1") {
		t.Fatal(err)
	}
	if !assertPluginIsEnabled(t, rabbitmqContainer, "rabbitmq_shovel", "rabbitmq_random_exchange") {
		t.Fatal(err)
	}
}

func assertEntity(t *testing.T, container testcontainers.Container, listCommand string, entities ...string) bool {
	t.Helper()

	ctx := context.Background()

	cmd := []string{"rabbitmqadmin", "list", listCommand}

	_, out, err := container.Exec(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}

	check, err := io.ReadAll(out)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entities {
		if !strings.Contains(string(check), e) {
			return false
		}
	}

	return true
}

func assertEntityWithVHost(t *testing.T, container testcontainers.Container, listCommand string, vhostID int, entities ...string) bool {
	t.Helper()

	ctx := context.Background()

	cmd := []string{"rabbitmqadmin", "list", listCommand}
	if vhostID > 0 {
		cmd = append(cmd, fmt.Sprintf("--vhost=vhost%d", vhostID))
	}

	_, out, err := container.Exec(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}

	check, err := io.ReadAll(out)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entities {
		if !strings.Contains(string(check), e) {
			return false
		}
	}

	return true
}

func assertPluginIsEnabled(t *testing.T, container testcontainers.Container, plugins ...string) bool {
	t.Helper()

	ctx := context.Background()

	for _, plugin := range plugins {

		_, out, err := container.Exec(ctx, []string{"rabbitmq-plugins", "is_enabled", plugin})
		if err != nil {
			t.Fatal(err)
		}

		check, err := io.ReadAll(out)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(check), plugin+" is enabled") {
			return false
		}
	}

	return true
}
