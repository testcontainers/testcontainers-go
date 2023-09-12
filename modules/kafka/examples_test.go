package kafka_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func ExampleRunContainer() {
	// runKafkaContainer {
	ctx := context.Background()

	kafkaContainer, err := kafka.RunContainer(ctx)
	if err != nil {
		panic(err)
	}

	// Clean up the container after
	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := kafkaContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
