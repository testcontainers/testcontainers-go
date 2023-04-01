package rabbitmq

import (
	"context"
	"fmt"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestRabbitMQ(t *testing.T) {
	ctx := context.Background()

	container, err := startContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions

	amqpConnection, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s", container.endpoint))

	if err != nil {
		t.Fatal(fmt.Errorf("error creating amqp client: %w", err))
	}

	err = amqpConnection.Close()
	if err != nil {
		t.Fatal(fmt.Errorf("error closing amqp connection: %s", err))
	}
}
