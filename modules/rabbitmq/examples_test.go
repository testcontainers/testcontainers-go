package rabbitmq_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func ExampleRunContainer() {
	// runRabbitMQContainer {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.RunContainer(ctx, testcontainers.WithImage("rabbitmq:3.12-management-alpine"))
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
