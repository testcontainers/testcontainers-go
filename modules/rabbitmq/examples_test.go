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

	// Clean up the container after
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

	fmt.Println(assertPluginIsEnabled(rabbitmqContainer, "rabbitmq_shovel"))
	fmt.Println(assertPluginIsEnabled(rabbitmqContainer, "rabbitmq_random_exchange"))

	// Output:
	// true
	// true
}

func assertPluginIsEnabled(container testcontainers.Container, plugin string) bool {
	ctx := context.Background()

	_, out, err := container.Exec(ctx, []string{"rabbitmq-plugins", "is_enabled", plugin})
	if err != nil {
		panic(err)
	}

	check, err := io.ReadAll(out)
	if err != nil {
		panic(err)
	}

	return strings.Contains(string(check), plugin+" is enabled")
}
