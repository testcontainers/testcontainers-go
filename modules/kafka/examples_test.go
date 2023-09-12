package kafka_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func ExampleRunContainer() {
	// runKafkaContainer {
	ctx := context.Background()

	kafkaContainer, err := kafka.RunContainer(ctx, testcontainers.WithImage("confluentinc/cp-kafka:7.3.3"))
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
