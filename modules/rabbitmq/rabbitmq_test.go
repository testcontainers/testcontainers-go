package rabbitmq_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func TestRunContainer_withAllSettings(t *testing.T) {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		// addVirtualHosts {
		rabbitmq.WithStartupCommand(rabbitmq.VirtualHost{Name: "vhost1"}),
		rabbitmq.WithStartupCommand(rabbitmq.VirtualHostLimit{VHost: "vhost1", Name: "max-connections", Value: 1}),
		rabbitmq.WithStartupCommand(rabbitmq.VirtualHost{Name: "vhost2", Tracing: true}),
		// }
		// addExchanges {
		rabbitmq.WithStartupCommand(rabbitmq.Exchange{Name: "direct-exchange", Type: "direct"}),
		rabbitmq.WithStartupCommand(rabbitmq.Exchange{
			Name: "topic-exchange",
			Type: "topic",
		}),
		rabbitmq.WithStartupCommand(rabbitmq.Exchange{
			VHost:      "vhost1",
			Name:       "topic-exchange-2",
			Type:       "topic",
			AutoDelete: false,
			Internal:   false,
			Durable:    true,
			Args:       map[string]interface{}{},
		}),
		rabbitmq.WithStartupCommand(rabbitmq.Exchange{
			VHost: "vhost2",
			Name:  "topic-exchange-3",
			Type:  "topic",
		}),
		rabbitmq.WithStartupCommand(rabbitmq.Exchange{
			Name:       "topic-exchange-4",
			Type:       "topic",
			AutoDelete: false,
			Internal:   false,
			Durable:    true,
			Args:       map[string]interface{}{},
		}),
		// }
		// addQueues {
		rabbitmq.WithStartupCommand(rabbitmq.Queue{Name: "queue1"}),
		rabbitmq.WithStartupCommand(rabbitmq.Queue{
			Name:       "queue2",
			AutoDelete: true,
			Durable:    false,
			Args:       map[string]interface{}{"x-message-ttl": 1000},
		}),
		rabbitmq.WithStartupCommand(rabbitmq.Queue{
			VHost:      "vhost1",
			Name:       "queue3",
			AutoDelete: true,
			Durable:    false,
			Args:       map[string]interface{}{"x-message-ttl": 1000},
		}),
		rabbitmq.WithStartupCommand(rabbitmq.Queue{VHost: "vhost2", Name: "queue4"}),
		// }
		// addBindings {
		rabbitmq.WithStartupCommand(rabbitmq.NewBinding("direct-exchange", "queue1")),
		rabbitmq.WithStartupCommand(rabbitmq.NewBindingWithVHost("vhost1", "topic-exchange-2", "queue3")),
		rabbitmq.WithStartupCommand(rabbitmq.Binding{
			VHost:           "vhost2",
			Source:          "topic-exchange-3",
			Destination:     "queue4",
			RoutingKey:      "ss7",
			DestinationType: "queue",
			Args:            map[string]interface{}{},
		}),
		// }
		// addUsers {
		rabbitmq.WithStartupCommand(rabbitmq.User{
			Name:     "user1",
			Password: "password1",
		}),
		rabbitmq.WithStartupCommand(rabbitmq.User{
			Name:     "user2",
			Password: "password2",
			Tags:     []string{"administrator"},
		}),
		// }
		// addPermissions {
		rabbitmq.WithStartupCommand(rabbitmq.NewPermission("vhost1", "user1", ".*", ".*", ".*")),
		// }
		// addPolicies {
		rabbitmq.WithStartupCommand(rabbitmq.Policy{
			Name:       "max length policy",
			Pattern:    "^dog",
			Definition: map[string]interface{}{"max-length": 1},
			Priority:   1,
			ApplyTo:    "queues",
		}),
		rabbitmq.WithStartupCommand(rabbitmq.Policy{
			Name:       "alternate exchange policy",
			Pattern:    "^direct-exchange",
			Definition: map[string]interface{}{"alternate-exchange": "amq.direct"},
		}),
		rabbitmq.WithStartupCommand(rabbitmq.Policy{
			VHost:   "vhost2",
			Name:    "ha-all",
			Pattern: ".*",
			Definition: map[string]interface{}{
				"ha-mode":      "all",
				"ha-sync-mode": "automatic",
			},
		}),
		rabbitmq.WithStartupCommand(rabbitmq.OperatorPolicy{
			Name:       "operator policy 1",
			Pattern:    "^queue1",
			Definition: map[string]interface{}{"message-ttl": 1000},
			Priority:   1,
			ApplyTo:    "queues",
		}),
		// }
		// enablePlugins {
		rabbitmq.WithStartupCommand(rabbitmq.Plugin("rabbitmq_shovel"), rabbitmq.Plugin("rabbitmq_random_exchange")),
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

	if !assertEntity(rabbitmqContainer, "queues", "queue1", "queue2", "queue3", "queue4") {
		t.Fatal(err)
	}
	if !assertEntity(rabbitmqContainer, "exchanges", "direct-exchange", "topic-exchange", "topic-exchange-2", "topic-exchange-3", "topic-exchange-4") {
		t.Fatal(err)
	}
	if !assertEntity(rabbitmqContainer, "users", "user1", "user2") {
		t.Fatal(err)
	}
	if !assertEntity(rabbitmqContainer, "policies", "max length policy", "alternate exchange policy") {
		t.Fatal(err)
	}
	if !assertEntityWithVHost(rabbitmqContainer, "policies", 2, "max length policy", "alternate exchange policy") {
		t.Fatal(err)
	}
	if !assertEntity(rabbitmqContainer, "operator_policies", "operator policy 1") {
		t.Fatal(err)
	}
	if !assertPluginIsEnabled(rabbitmqContainer, "rabbitmq_shovel", "rabbitmq_random_exchange") {
		t.Fatal(err)
	}
}
