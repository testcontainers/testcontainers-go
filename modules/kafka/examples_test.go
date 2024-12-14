package kafka_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/IBM/sarama"
	"github.com/docker/docker/api/types/container"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRun() {
	// runKafkaContainer {
	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		kafka.WithClusterID("test-cluster"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
			log.Printf("failed to terminate container: %v", err)
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

func ExampleKafkaContainer_BrokersByHostDockerInternal() {
	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		kafka.WithClusterID("test-cluster"),
	)
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// Clean up the container after
	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	state, err := kafkaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(kafkaContainer.ClusterID)
	fmt.Println(state.Running)

	const topic = "example-topic"

	// Produce a message from the host that will be read by a consumer in another docker container
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		log.Print(err)
		return
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Print(err)
		return
	}

	if _, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder("example_message_value"),
	}); err != nil {
		log.Print(err)
		return
	}

	// getBrokersByHostDockerInternal {
	brokers, err = kafkaContainer.BrokersByHostDockerInternal(ctx)
	if err != nil {
		log.Print(err)
		return
	}

	// Run another container that can connect to the kafka container via hostname "host.docker.internal"
	kcat, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:      "confluentinc/cp-kafkacat",
				Entrypoint: []string{"kafkacat"},
				Cmd:        []string{"-b", strings.Join(brokers, ","), "-C", "-t", topic, "-c", "1"},
				WaitingFor: wait.ForExit(),

				// Add host.docker.internal to the consumer container so it can contact the kafka borkers
				HostConfigModifier: func(hc *container.HostConfig) {
					hc.ExtraHosts = append(hc.ExtraHosts, "host.docker.internal:host-gateway")
				},
			},
			Started: true,
		},
	)
	if err != nil {
		log.Printf("kafkacat error: %v", err)
		return
	}

	lr, err := kcat.Logs(ctx)
	if err != nil {
		log.Printf("kafkacat logs error: %v", err)
		return
	}

	logs, err := io.ReadAll(lr)
	if err != nil {
		log.Printf("kafkacat logs read error: %v", err)
		return
	}

	fmt.Println("read message:", string(bytes.TrimSpace(logs)))
	// }

	// Output:
	// test-cluster
	// true
	// read message: example_message_value
}

func ExampleKafkaContainer_BrokersByContainerName() {
	ctx := context.Background()

	// getBrokersByContainerName_Kafka {
	net, err := network.New(ctx)
	if err != nil {
		log.Printf("failed to create network: %s", err)
		return
	}

	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		kafka.WithClusterID("test-cluster"),
		network.WithNetwork(nil, net), // Run kafka test container in a new docker network
	)
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	// Clean up the container after
	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	state, err := kafkaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(kafkaContainer.ClusterID)
	fmt.Println(state.Running)

	const topic = "example-topic"

	// Produce a message from the host that will be read by a consumer in another docker container
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		log.Print(err)
		return
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Print(err)
		return
	}

	if _, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder("example_message_value"),
	}); err != nil {
		log.Print(err)
		return
	}

	// getBrokersByContainerName_Kcat {
	brokers, err = kafkaContainer.BrokersByContainerName(ctx)
	if err != nil {
		log.Print(err)
		return
	}

	// Run another container that can connect to the kafka container via the kafka containers name
	kcat, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:      "confluentinc/cp-kafkacat",
				Entrypoint: []string{"kafkacat"},
				Cmd:        []string{"-b", strings.Join(brokers, ","), "-C", "-t", topic, "-c", "1"},
				WaitingFor: wait.ForExit(),
				Networks:   []string{net.Name}, // Run kafkacat in the same docker network as the testcontainer
			},
			Started: true,
		},
	)
	if err != nil {
		log.Printf("kafkacat error: %v", err)
		return
	}

	lr, err := kcat.Logs(ctx)
	if err != nil {
		log.Printf("kafkacat logs error: %v", err)
		return
	}

	logs, err := io.ReadAll(lr)
	if err != nil {
		log.Printf("kafkacat logs read error: %v", err)
		return
	}

	fmt.Println("read message:", string(bytes.TrimSpace(logs)))
	// }

	// Output:
	// test-cluster
	// true
	// read message: example_message_value
}

func ExampleKafkaContainer_BrokersByContainerId() {
	ctx := context.Background()

	net, err := network.New(ctx)
	if err != nil {
		log.Printf("failed to create network: %s", err)
		return
	}

	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		kafka.WithClusterID("test-cluster"),
		network.WithNetwork(nil, net), // Run kafka test container in a new docker network
	)
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// Clean up the container after
	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	state, err := kafkaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(kafkaContainer.ClusterID)
	fmt.Println(state.Running)

	const topic = "example-topic"

	// Produce a message from the host that will be read by a consumer in another docker container
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		log.Print(err)
		return
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Print(err)
		return
	}

	if _, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder("example_message_value"),
	}); err != nil {
		log.Print(err)
		return
	}

	brokers, err = kafkaContainer.BrokersByContainerId(ctx)
	if err != nil {
		log.Print(err)
		return
	}

	// Run another container that can connect to the kafka container via the kafka containers ContainerID
	kcat, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:      "confluentinc/cp-kafkacat",
				Entrypoint: []string{"kafkacat"},
				Cmd:        []string{"-b", strings.Join(brokers, ","), "-C", "-t", topic, "-c", "1"},
				WaitingFor: wait.ForExit(),
				Networks:   []string{net.Name}, // Run kafkacat in the same docker network as the testcontainer
			},
			Started: true,
		},
	)
	if err != nil {
		log.Printf("kafkacat error: %v", err)
		return
	}

	lr, err := kcat.Logs(ctx)
	if err != nil {
		log.Printf("kafkacat logs error: %v", err)
		return
	}

	logs, err := io.ReadAll(lr)
	if err != nil {
		log.Printf("kafkacat logs read error: %v", err)
		return
	}

	fmt.Println("read message:", string(bytes.TrimSpace(logs)))

	// Output:
	// test-cluster
	// true
	// read message: example_message_value
}
