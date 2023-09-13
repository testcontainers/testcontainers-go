package rabbitmq_test

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func ExampleRunContainer() {
	// runRabbitMQContainer {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		rabbitmq.WithAdminPassword("password"),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := rabbitmqContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_withSSL() {
	// enableSSL {
	ctx := context.Background()

	serverCa := filepath.Join("testdata", "certs", "server_ca.pem")

	sslSettings := rabbitmq.SSLSettings{
		CACertFile:        serverCa,
		CertFile:          filepath.Join("testdata", "certs", "server_cert.pem"),
		KeyFile:           filepath.Join("testdata", "certs", "server_key.pem"),
		VerificationMode:  rabbitmq.SSLVerificationModePeer,
		FailIfNoCert:      true,
		VerificationDepth: 1,
	}

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		rabbitmq.WithSSL(sslSettings),
	)
	if err != nil {
		panic(err)
	}
	// }

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	state, err := rabbitmqContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_withPlugins() {
	// enablePlugins {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		rabbitmq.WithPluginsEnabled("rabbitmq_shovel", "rabbitmq_random_exchange"),
	)
	if err != nil {
		panic(err)
	}
	// }

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	fmt.Println(assertPluginIsEnabled(rabbitmqContainer, "rabbitmq_shovel", "rabbitmq_random_exchange"))

	// Output:
	// true
}

func ExampleRunContainer_withAllSettings() {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		rabbitmq.WithVirtualHost(rabbitmq.VirtualHost{Name: "vhost1"}),
		rabbitmq.WithVirtualHostLimit(rabbitmq.VirtualHostLimit{VHost: "vhost1", Name: "max-connections", Value: 1}),
		rabbitmq.WithVirtualHost(rabbitmq.VirtualHost{Name: "vhost2", Tracing: true}),

		rabbitmq.WithExchange(rabbitmq.Exchange{Name: "direct-exchange", Type: "direct"}),
		rabbitmq.WithExchange(rabbitmq.Exchange{
			Name: "topic-exchange",
			Type: "topic",
		}),
		rabbitmq.WithExchange(rabbitmq.Exchange{
			VHost:      "vhost1",
			Name:       "topic-exchange-2",
			Type:       "topic",
			AutoDelete: false,
			Internal:   false,
			Durable:    true,
			Args:       map[string]interface{}{},
		}),
		rabbitmq.WithExchange(rabbitmq.Exchange{
			VHost: "vhost2",
			Name:  "topic-exchange-3",
			Type:  "topic",
		}),
		rabbitmq.WithExchange(rabbitmq.Exchange{
			Name:       "topic-exchange-4",
			Type:       "topic",
			AutoDelete: false,
			Internal:   false,
			Durable:    true,
			Args:       map[string]interface{}{},
		}),

		rabbitmq.WithQueue(rabbitmq.Queue{Name: "queue1"}),
		rabbitmq.WithQueue(rabbitmq.Queue{
			Name:       "queue2",
			AutoDelete: true,
			Durable:    false,
			Args:       map[string]interface{}{"x-message-ttl": 1000},
		}),
		rabbitmq.WithQueue(rabbitmq.Queue{
			VHost:      "vhost1",
			Name:       "queue3",
			AutoDelete: true,
			Durable:    false,
			Args:       map[string]interface{}{"x-message-ttl": 1000},
		}),
		rabbitmq.WithQueue(rabbitmq.Queue{VHost: "vhost2", Name: "queue4"}),

		rabbitmq.WithBinding(rabbitmq.NewBinding("direct-exchange", "queue1")),
		rabbitmq.WithBinding(rabbitmq.NewBindingWithVHost("vhost1", "topic-exchange-2", "queue3")),
		rabbitmq.WithBinding(rabbitmq.Binding{
			VHost:           "vhost2",
			Source:          "topic-exchange-3",
			Destination:     "queue4",
			RoutingKey:      "ss7",
			DestinationType: "queue",
			Args:            map[string]interface{}{},
		}),

		rabbitmq.WithUser(rabbitmq.User{
			Name:     "user1",
			Password: "password1",
		}),
		rabbitmq.WithUser(rabbitmq.User{
			Name:     "user2",
			Password: "password2",
			Tags:     []string{"administrator"},
		}),

		rabbitmq.WithPermission(rabbitmq.NewPermission("vhost1", "user1", ".*", ".*", ".*")),

		rabbitmq.WithPolicy(rabbitmq.Policy{
			Name:       "max length policy",
			Pattern:    "^dog",
			Definition: map[string]interface{}{"max-length": 1},
			Priority:   1,
			ApplyTo:    "queues",
		}),
		rabbitmq.WithPolicy(rabbitmq.Policy{
			Name:       "alternate exchange policy",
			Pattern:    "^direct-exchange",
			Definition: map[string]interface{}{"alternate-exchange": "amq.direct"},
		}),
		rabbitmq.WithPolicy(rabbitmq.Policy{
			VHost:   "vhost2",
			Name:    "ha-all",
			Pattern: ".*",
			Definition: map[string]interface{}{
				"ha-mode":      "all",
				"ha-sync-mode": "automatic",
			},
		}),

		rabbitmq.WithOperatorPolicy(rabbitmq.OperatorPolicy{
			Name:       "operator policy 1",
			Pattern:    "^queue1",
			Definition: map[string]interface{}{"message-ttl": 1000},
			Priority:   1,
			ApplyTo:    "queues",
		}),

		rabbitmq.WithPluginsEnabled("rabbitmq_shovel", "rabbitmq_random_exchange"),
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	fmt.Println(assertEntity(rabbitmqContainer, "queues", "queue1", "queue2", "queue3", "queue4"))
	fmt.Println(assertEntity(rabbitmqContainer, "exchanges", "direct-exchange", "topic-exchange", "topic-exchange-2", "topic-exchange-3", "topic-exchange-4"))
	fmt.Println(assertEntity(rabbitmqContainer, "users", "user1", "user2"))
	fmt.Println(assertEntity(rabbitmqContainer, "policies", "max length policy", "alternate exchange policy"))
	fmt.Println(assertEntityWithVHost(rabbitmqContainer, "policies", 2, "max length policy", "alternate exchange policy"))
	fmt.Println(assertEntity(rabbitmqContainer, "operator_policies", "operator policy 1"))
	fmt.Println(assertPluginIsEnabled(rabbitmqContainer, "rabbitmq_shovel", "rabbitmq_random_exchange"))

	// Output:
	// true
	// true
	// true
	// true
	// true
	// true
	// true
}

func assertEntity(container testcontainers.Container, listCommand string, entities ...string) bool {
	ctx := context.Background()

	cmd := []string{"rabbitmqadmin", "list", listCommand}

	_, out, err := container.Exec(ctx, cmd)
	if err != nil {
		panic(err)
	}

	check, err := io.ReadAll(out)
	if err != nil {
		panic(err)
	}

	for _, e := range entities {
		if !strings.Contains(string(check), e) {
			return false
		}
	}

	return true
}

func assertEntityWithVHost(container testcontainers.Container, listCommand string, vhostID int, entities ...string) bool {
	ctx := context.Background()

	cmd := []string{"rabbitmqadmin", "list", listCommand}
	if vhostID > 0 {
		cmd = append(cmd, fmt.Sprintf("--vhost=vhost%d", vhostID))
	}

	_, out, err := container.Exec(ctx, cmd)
	if err != nil {
		panic(err)
	}

	check, err := io.ReadAll(out)
	if err != nil {
		panic(err)
	}

	for _, e := range entities {
		if !strings.Contains(string(check), e) {
			return false
		}
	}

	return true
}

func assertPluginIsEnabled(container testcontainers.Container, plugins ...string) bool {
	ctx := context.Background()

	for _, plugin := range plugins {

		_, out, err := container.Exec(ctx, []string{"rabbitmq-plugins", "is_enabled", plugin})
		if err != nil {
			panic(err)
		}

		check, err := io.ReadAll(out)
		if err != nil {
			panic(err)
		}

		if !strings.Contains(string(check), plugin+" is enabled") {
			return false
		}
	}

	return true
}
