package examples_test

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func Example_connectUsingAmqp() {
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
