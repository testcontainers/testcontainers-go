package rabbitmq_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mdelapenya/tlscert"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func ExampleRunContainer() {
	// runRabbitMQContainer {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.12.11-management-alpine"),
		rabbitmq.WithAdminUsername("admin"),
		rabbitmq.WithAdminPassword("password"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := rabbitmqContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
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
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	amqpURL, err := rabbitmqContainer.AmqpURL(ctx)
	if err != nil {
		log.Fatalf("failed to get AMQP URL: %s", err) // nolint:gocritic
	}

	amqpConnection, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %s", err)
	}
	defer func() {
		err := amqpConnection.Close()
		if err != nil {
			log.Fatalf("failed to close connection: %s", err)
		}
	}()

	fmt.Println(amqpConnection.IsClosed())

	// Output:
	// false
}

func ExampleRunContainer_withSSL() {
	// enableSSL {
	ctx := context.Background()

	tmpDir := os.TempDir()
	certDirs := tmpDir + "/rabbitmq"
	if err := os.MkdirAll(certDirs, 0755); err != nil {
		log.Fatalf("failed to create temporary directory: %s", err)
	}
	defer os.RemoveAll(certDirs)

	// generates the CA certificate and the certificate
	caCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:      "ca",
		Host:      "localhost,127.0.0.1",
		IsCA:      true,
		ParentDir: certDirs,
	})
	if caCert == nil {
		log.Fatal("failed to generate CA certificate")
	}

	cert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:      "client",
		Host:      "localhost,127.0.0.1",
		IsCA:      true,
		Parent:    caCert,
		ParentDir: certDirs,
	})
	if cert == nil {
		log.Fatal("failed to generate certificate")
	}

	sslSettings := rabbitmq.SSLSettings{
		CACertFile:        caCert.CertPath,
		CertFile:          cert.CertPath,
		KeyFile:           cert.KeyPath,
		VerificationMode:  rabbitmq.SSLVerificationModePeer,
		FailIfNoCert:      true,
		VerificationDepth: 1,
	}

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		rabbitmq.WithSSL(sslSettings),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	// }

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := rabbitmqContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
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
		testcontainers.WithAfterReadyCommand(
			testcontainers.NewRawCommand([]string{"rabbitmq_shovel"}),
			testcontainers.NewRawCommand([]string{"rabbitmq_random_exchange"}),
		),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
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
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := rabbitmqContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	logs, err := rabbitmqContainer.Logs(ctx)
	if err != nil {
		log.Fatalf("failed to get logs: %s", err) // nolint:gocritic
	}

	bytes, err := io.ReadAll(logs)
	if err != nil {
		log.Fatalf("failed to read logs: %s", err)
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
			log.Fatalf("failed to execute command: %s", err)
		}

		check, err := io.ReadAll(out)
		if err != nil {
			log.Fatalf("failed to read output: %s", err)
		}

		if !strings.Contains(string(check), plugin+" is enabled") {
			return false
		}
	}

	return true
}
