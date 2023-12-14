package rabbitmq_test

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func ExampleRunContainer() {
	// runRabbitMQContainer {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		rabbitmq.WithAdminUsername("admin"),
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

func ExampleRunContainer_connectUsingAmqp() {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		rabbitmq.WithAdminUsername("admin"),
		rabbitmq.WithAdminPassword("password"),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	amqpURL, err := rabbitmqContainer.AmqpURL(ctx)
	if err != nil {
		panic(err)
	}

	amqpConnection, err := amqp.Dial(amqpURL)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := amqpConnection.Close()
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println(amqpConnection.IsClosed())

	// Output:
	// false
}

func ExampleRunContainer_withSSL() {
	// enableSSL {
	ctx := context.Background()

	sslSettings := rabbitmq.SSLSettings{
		CACertFile:        filepath.Join("testdata", "certs", "server_ca.pem"),
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
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		// Multiple test implementations of the Executable interface, specific to RabbitMQ, exist in the types_test.go file.
		// Please refer to them for more examples.
		testcontainers.WithStartupCommand(
			testcontainers.NewRawCommand([]string{"rabbitmq_shovel"}),
			testcontainers.NewRawCommand([]string{"rabbitmq_random_exchange"}),
		),
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	fmt.Println(assertPlugins(rabbitmqContainer, "rabbitmq_shovel", "rabbitmq_random_exchange"))

	// Output:
	// true
}

func ExampleRunContainer_withCustomConfigFile() {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	logs, err := rabbitmqContainer.Logs(ctx)
	if err != nil {
		panic(err)
	}

	bytes, err := io.ReadAll(logs)
	if err != nil {
		panic(err)
	}

	fmt.Println(strings.Contains(string(bytes), "config file(s) : /etc/rabbitmq/rabbitmq-testcontainers.conf"))

	// Output:
	// true
}

func assertPlugins(container testcontainers.Container, plugins ...string) bool {
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
