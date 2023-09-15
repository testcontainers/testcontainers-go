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

	kafkaContainer, err := kafka.RunContainer(ctx,
		kafka.WithClusterID("test-cluster"),
		testcontainers.WithImage("confluentinc/confluent-local:7.5.0"),
	)
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

	fmt.Println(kafkaContainer.ClusterID)
	fmt.Println(state.Running)

	// Output:
	// test-cluster
	// true
}
