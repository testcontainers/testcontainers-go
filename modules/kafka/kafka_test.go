package kafka

import (
	"context"
	"strings"
	"testing"

	"github.com/IBM/sarama"

	"github.com/testcontainers/testcontainers-go"
)

func TestKafka(t *testing.T) {
	topic := "some-topic"

	ctx := context.Background()

	kafkaContainer, err := RunContainer(ctx, WithClusterID("kraftCluster"), testcontainers.WithImage("confluentinc/confluent-local:7.5.0"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

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

	_, err := RunContainer(ctx, WithClusterID("kraftCluster"), testcontainers.WithImage("confluentinc/confluent-local:6.3.3"))
	if err == nil {
		t.Fatal(err)
	}
}

func TestConfigureQuorumVoters(t *testing.T) {
	tests := []struct {
		name           string
		req            *testcontainers.GenericContainerRequest
		expectedVoters string
	}{
		{
			name: "voters on localhost",
			req: &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Env: map[string]string{},
				},
			},
			expectedVoters: "1@localhost:9094",
		},
		{
			name: "voters on first network alias of the first network",
			req: &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Env:      map[string]string{},
					Networks: []string{"foo", "bar", "baaz"},
					NetworkAliases: map[string][]string{
						"foo":  {"foo0", "foo1", "foo2", "foo3"},
						"bar":  {"bar0", "bar1", "bar2", "bar3"},
						"baaz": {"baaz0", "baaz1", "baaz2", "baaz3"},
					},
				},
			},
			expectedVoters: "1@foo0:9094",
		},
		{
			name: "voters on localhost if alias but no networks",
			req: &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					NetworkAliases: map[string][]string{
						"foo":  {"foo0", "foo1", "foo2", "foo3"},
						"bar":  {"bar0", "bar1", "bar2", "bar3"},
						"baaz": {"baaz0", "baaz1", "baaz2", "baaz3"},
					},
				},
			},
			expectedVoters: "1@localhost:9094",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configureControllerQuorumVoters(test.req)

			if test.req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"] != test.expectedVoters {
				t.Fatalf("expected KAFKA_CONTROLLER_QUORUM_VOTERS to be %s, got %s", test.expectedVoters, test.req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"])
			}
		})
	}
}

func TestValidateKRaftVersion(t *testing.T) {
	tests := []struct {
		name    string
		image   string
		wantErr bool
	}{
		{
			name:    "Official: valid version",
			image:   "confluentinc/confluent-local:7.5.0",
			wantErr: false,
		},
		{
			name:    "Official: valid, limit version",
			image:   "confluentinc/confluent-local:7.4.0",
			wantErr: false,
		},
		{
			name:    "Official: invalid, low version",
			image:   "confluentinc/confluent-local:7.3.99",
			wantErr: true,
		},
		{
			name:    "Official: invalid, too low version",
			image:   "confluentinc/confluent-local:5.0.0",
			wantErr: true,
		},
		{
			name:    "Unofficial does not validate KRaft version",
			image:   "my-kafka:1.0.0",
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateKRaftVersion(test.image)

			if test.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !test.wantErr && err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
		})
	}
}
