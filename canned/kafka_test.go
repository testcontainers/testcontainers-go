package canned_test

import (
	"fmt"
	"github.com/testcontainers/testcontainers-go/canned"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"reflect"
	"testing"
)

const (
	KAFKA_TOPIC = "myTopic"
)

func TestKafkaConsumerAndProducerUsingTestContainer(t *testing.T) {
	kafkaCluster := canned.NewKafkaCluster()
	kafkaCluster.StartCluster()
	kafkaServer := kafkaCluster.GetKafkaHost()
	producedMessages := []string{"Trying", "out", "kafka", "with", "test", "containers"}

	produceKafkaMessages(kafkaServer, producedMessages)
	consumedMessages := consumeKafkaMessages(kafkaServer)

	if !reflect.DeepEqual(producedMessages, consumedMessages) {
		t.Fatalf("Consumed messages are not equal to produced messages. [%s] != [%s]", consumedMessages, producedMessages)
	}
}

func produceKafkaMessages(kafkaServer string, messages []string) {
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServer,
	})
	if err != nil {
		panic(err)
	}
	defer kafkaProducer.Close()

	fmt.Printf("Producing messages into kafka...\n")

	topic := KAFKA_TOPIC
	for _, word := range messages {
		deliveryChan := make(chan kafka.Event)

		kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
			Key:            []byte("key"),
		}, deliveryChan)

		e := <-deliveryChan
		m := e.(*kafka.Message)

		if m.TopicPartition.Error != nil {
			fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
		} else {
			fmt.Printf("Delivered message [%s] to topic %s [%d] at offset %v\n",
				word, *m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
		}
	}
}

func consumeKafkaMessages(kafkaServer string) []string {
	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServer,
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}
	err = kafkaConsumer.SubscribeTopics([]string{KAFKA_TOPIC}, nil)
	if err != nil {
		panic(err)
	}
	defer kafkaConsumer.Close()

	fmt.Printf("Consuming messages from kafka...\n")

	var consumedMessages []string
	run := true
	for run == true {
		select {
		default:
			ev := kafkaConsumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("Message on %s: %s\n", e.TopicPartition, string(e.Value))
				consumedMessages = append(consumedMessages, string(e.Value))
			default:
				fmt.Printf("Consumed all messges. Stopping consumer.\n")
				run = false
			}
		}
	}

	return consumedMessages
}
