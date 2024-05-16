package kafka_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/IBM/sarama"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestKafka(t *testing.T) {
	topic := "some-topic"

	ctx := context.Background()

	kafkaContainer, err := kafka.RunContainer(ctx, kafka.WithClusterID("kraftCluster"), testcontainers.WithImage("confluentinc/confluent-local:7.5.0"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertAdvertisedListeners(t, kafkaContainer)

	if !strings.EqualFold(kafkaContainer.ClusterID, "kraftCluster") {
		t.Fatalf("expected clusterID to be %s, got %s", "kraftCluster", kafkaContainer.ClusterID)
	}

	// getBrokers {
	brokers, err := kafkaContainer.Brokers(ctx)
	// }
	if err != nil {
		t.Fatal(err)
	}

	config := sarama.NewConfig()
	client, err := sarama.NewConsumerGroup(brokers, "groupName", config)
	if err != nil {
		t.Fatal(err)
	}

	consumer, ready, done, cancel := NewTestKafkaConsumer(t)
	go func() {
		if err := client.Consume(context.Background(), []string{topic}, consumer); err != nil {
			cancel()
		}
	}()

	// wait for the consumer to be ready
	<-ready

	// perform assertions

	// set config to true because successfully delivered messages will be returned on the Successes channel
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		cancel()
		t.Fatal(err)
	}

	if _, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder("value"),
	}); err != nil {
		cancel()
		t.Fatal(err)
	}

	<-done

	if !strings.EqualFold(string(consumer.message.Key), "key") {
		t.Fatalf("expected key to be %s, got %s", "key", string(consumer.message.Key))
	}
	if !strings.EqualFold(string(consumer.message.Value), "value") {
		t.Fatalf("expected value to be %s, got %s", "value", string(consumer.message.Value))
	}
}

func TestKafka_invalidVersion(t *testing.T) {
	ctx := context.Background()

	_, err := kafka.RunContainer(ctx, kafka.WithClusterID("kraftCluster"), testcontainers.WithImage("confluentinc/confluent-local:6.3.3"))
	if err == nil {
		t.Fatal(err)
	}
}

func TestKafka_network(t *testing.T) {
	ctx := context.Background()
	var err error

	const (
		topic_in  = "topic_in"
		topic_out = "topic_out"
	)

	Network, err := network.New(ctx, network.WithCheckDuplicate())
	if err != nil {
		t.Fatal(err)
	}

	// kafkaWithListener {
	KafkaContainer, err := kafka.RunContainer(ctx,
		kafka.WithClusterID("test-cluster"),
		testcontainers.WithImage("confluentinc/confluent-local:7.6.1"),
		network.WithNetwork([]string{"kafka"}, Network),
		kafka.WithListener([]kafka.KafkaListener{
			{
				Name: "INTERNAL",
				Ip:   "kafka",
				Port: "9092",
			},
		}),
	)
	// }

	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	brokers, err := KafkaContainer.Brokers(context.TODO())
	if err != nil {
		t.Fatal("failed to get brokers", err)
	}

	createTopics(brokers, []string{topic_in, topic_out})

	initKafkaTest(ctx, Network.Name, "kafka:9092", topic_in, topic_out)

	// perform assertions

	// set config to true because successfully delivered messages will be returned on the Successes channel
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		t.Fatal(err)
	}

	// Act
	key := "wow"
	text_msg := "test-input-external"

	if _, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic_in,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(text_msg),
	}); err != nil {
		t.Fatal(err)
	}

	client, err := sarama.NewConsumerGroup(brokers, "groupName", config)
	if err != nil {
		t.Fatal(err)
	}

	consumer, _, done, cancel := NewTestKafkaConsumer(t)
	go func() {
		if err := client.Consume(context.Background(), []string{topic_out}, consumer); err != nil {
			cancel()
		}
	}()

	// wait for the consumer to be ready
	<-done

	if consumer.message == nil {
		t.Fatal("Empty message")
	}

	// Assert
	if !strings.Contains(string(consumer.message.Value), text_msg) {
		t.Error("got wrong string")
	}
}

func initKafkaTest(ctx context.Context, network string, brokers string, input string, output string) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:       "./docker",
			Dockerfile:    "Dockerfile",
			PrintBuildLog: true,
			KeepImage:     true,
		},
		WaitingFor: wait.ForLog("start consuming events"),
		Env: map[string]string{
			"KAFKA_BROKERS":   brokers,
			"KAFKA_TOPIC_IN":  input,
			"KAFKA_TOPIC_OUT": output,
		},
		Networks: []string{network},
	}

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func createTopics(brokers []string, topics []string) error {
	t := &sarama.CreateTopicsRequest{}
	t.TopicDetails = make(map[string]*sarama.TopicDetail, len(topics))
	for _, elem := range topics {
		t.TopicDetails[elem] = &sarama.TopicDetail{NumPartitions: 1}
	}

	var err error

	c, err := sarama.NewClient(brokers, sarama.NewConfig())

	_, err = c.Brokers()[0].CreateTopics(t)
	if err != nil {
		return fmt.Errorf("failed to create topics: %w", err)
	}

	fmt.Println("succesfully created topics")

	return nil
}

// assertAdvertisedListeners checks that the advertised listeners are set correctly:
// - The INTERNAL:// protocol is using the hostname of the Kafka container
func assertAdvertisedListeners(t *testing.T, container testcontainers.Container) {
	hostname, err := container.Host(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	code, r, err := container.Exec(context.Background(), []string{"cat", "/usr/sbin/testcontainers_start.sh"})
	if err != nil {
		t.Fatal(err)
	}

	if code != 0 {
		t.Fatalf("expected exit code to be 0, got %d", code)
	}

	bs, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(bs), "INTERNAL://"+hostname+":9092") {
		t.Fatalf("expected advertised listeners to contain %s, got %s", "INTERNAL://"+hostname+":9092", string(bs))
	}
}
