package kafka_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRun_confluentinc() {
	// runKafkaContainerConfluentinc {
	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		kafka.WithClusterID("test-cluster"),
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

func ExampleRun_apacheNative() {
	// runKafkaContainerApacheNative {
	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx,
		"apache/kafka-native:4.0.1",
		kafka.WithClusterID("test-cluster"),
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

func ExampleRun_apacheNotNative() {
	// runKafkaContainerApacheNotNative {
	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx,
		"apache/kafka:4.0.1",
		kafka.WithClusterID("test-cluster"),
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

func ExampleRun_apacheNative_withApacheFlavor() {
	// runKafkaContainerWithApacheFlavor {
	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx,
		// the image might be different, for example
		// custom-registry/apache/kafka-native:4.0.1,
		// in which case the starter script would not
		// be correctly inferred, and should be overridden
		"apache/kafka-native:4.0.1",
		kafka.WithClusterID("test-cluster"),
		// this explicitly sets the starter script to use
		// the one compatible with Apache images
		kafka.WithApacheFlavor(),
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

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

func ExampleRun_confluentinc_withConfluentFlavor() {
	// runKafkaContainerWithConfluentFlavor {
	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx,
		// the image might be different, for example
		// custom-registry/confluentinc/confluent-local:7.5.0,
		// in which case the starter script might not
		// be correctly inferred, and should be overridden
		"confluentinc/confluent-local:7.5.0",
		kafka.WithClusterID("test-cluster"),
		// this explicitly sets the starter script to use
		// the one compatible with Confluent images
		kafka.WithConfluentFlavor(),
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

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

func ExampleRun_usingLocalhostListener() {
	ctx := context.Background()

	// runKafkaContainerAndUseLocalhostListener {
	kafkaContainer, err := kafka.Run(ctx, "apache/kafka:4.0.1",
		testcontainers.WithWaitStrategy(
			wait.NewExecStrategy([]string{
				"/opt/kafka/bin/kafka-topics.sh",
				"--bootstrap-server",
				"localhost:9095",
				"--list",
			}).
				WithExitCode(0).
				WithPollInterval(2*time.Second).
				WithStartupTimeout(120*time.Second),
		),
		kafka.WithClusterID("test-cluster"),
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

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
