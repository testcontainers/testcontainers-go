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

func ExampleRun() {
	// runRabbitMQContainer {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:3.12.11-management-alpine",
		rabbitmq.WithAdminUsername("admin"),
		rabbitmq.WithAdminPassword("password"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(rabbitmqContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := rabbitmqContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connectUsingAmqp() {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:3.7.25-management-alpine",
		rabbitmq.WithAdminUsername("admin"),
		rabbitmq.WithAdminPassword("password"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(rabbitmqContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	amqpURL, err := rabbitmqContainer.AmqpURL(ctx)
	if err != nil {
		log.Printf("failed to get AMQP URL: %s", err)
		return
	}

	amqpConnection, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Printf("failed to connect to RabbitMQ: %s", err)
		return
	}
	defer func() {
		err := amqpConnection.Close()
		if err != nil {
			log.Printf("failed to close connection: %s", err)
		}
	}()

	fmt.Println(amqpConnection.IsClosed())

	// Output:
	// false
}

func ExampleRun_withSSL() {
	// enableSSL {
	ctx := context.Background()

	tmpDir := os.TempDir()
	certDirs := tmpDir + "/rabbitmq"
	if err := os.MkdirAll(certDirs, 0o755); err != nil {
		log.Printf("failed to create temporary directory: %s", err)
		return
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
		log.Print("failed to generate CA certificate")
		return
	}

	cert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:      "client",
		Host:      "localhost,127.0.0.1",
		IsCA:      true,
		Parent:    caCert,
		ParentDir: certDirs,
	})
	if cert == nil {
		log.Print("failed to generate certificate")
		return
	}

	sslSettings := rabbitmq.SSLSettings{
		CACertFile:        caCert.CertPath,
		CertFile:          cert.CertPath,
		KeyFile:           cert.KeyPath,
		VerificationMode:  rabbitmq.SSLVerificationModePeer,
		FailIfNoCert:      true,
		VerificationDepth: 1,
	}

	rabbitmqContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:3.7.25-management-alpine",
		rabbitmq.WithSSL(sslSettings),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(rabbitmqContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := rabbitmqContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_withPlugins() {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:3.7.25-management-alpine",
		// Multiple test implementations of the Executable interface, specific to RabbitMQ, exist in the types_test.go file.
		// Please refer to them for more examples.
		testcontainers.WithAfterReadyCommand(
			testcontainers.NewRawCommand([]string{"rabbitmq_shovel"}),
			testcontainers.NewRawCommand([]string{"rabbitmq_random_exchange"}),
		),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(rabbitmqContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	if err = assertPlugins(rabbitmqContainer, "rabbitmq_shovel", "rabbitmq_random_exchange"); err != nil {
		log.Printf("failed to find plugin: %s", err)
		return
	}

	fmt.Println(true)

	// Output:
	// true
}

func ExampleRun_withCustomConfigFile() {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:3.7.25-management-alpine",
	)
	defer func() {
		if err := testcontainers.TerminateContainer(rabbitmqContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	logs, err := rabbitmqContainer.Logs(ctx)
	if err != nil {
		log.Printf("failed to get logs: %s", err)
		return
	}

	bytes, err := io.ReadAll(logs)
	if err != nil {
		log.Printf("failed to read logs: %s", err)
		return
	}

	fmt.Println(strings.Contains(string(bytes), "config file(s) : /etc/rabbitmq/rabbitmq-testcontainers.conf"))

	// Output:
	// true
}

func assertPlugins(container testcontainers.Container, plugins ...string) error {
	ctx := context.Background()

	for _, plugin := range plugins {
		_, out, err := container.Exec(ctx, []string{"rabbitmq-plugins", "is_enabled", plugin})
		if err != nil {
			return fmt.Errorf("failed to execute command: %w", err)
		}

		check, err := io.ReadAll(out)
		if err != nil {
			return fmt.Errorf("failed to read output: %w", err)
		}

		if !strings.Contains(string(check), plugin+" is enabled") {
			return fmt.Errorf("plugin %q is not enabled", plugin)
		}
	}

	return nil
}
