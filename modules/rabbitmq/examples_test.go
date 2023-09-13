package rabbitmq_test

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func ExampleRunContainer() {
	// runRabbitMQContainer {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx, testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"))
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
	ctx := context.Background()

	serverCa := filepath.Join("testdata", "certs", "server_ca.pem")

	sslSettings := rabbitmq.SSLSettings{
		CACertFile:       serverCa,
		CertFile:         filepath.Join("testdata", "certs", "server_cert.pem"),
		KeyFile:          filepath.Join("testdata", "certs", "server_key.pem"),
		VerificationMode: rabbitmq.SSLVerificationModePeer,
		FailIfNoCert:     true,
	}

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		rabbitmq.WithSSL(sslSettings),
	)
	if err != nil {
		panic(err)
	}

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
