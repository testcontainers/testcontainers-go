package kafka_native_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka_native"
)

func ExampleRun() {
	// runKafkaContainer {
	ctx := context.Background()

	kafkaContainer, err := kafka_native.Run(ctx,
		"apache/kafka-native:3.9.1",
		kafka_native.WithClusterID("test-cluster"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := kafkaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(kafkaContainer.ClusterID)
	fmt.Println(state.Running)

	// Output:
	// test-cluster
	// true
}
