package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	client = "kafka_rw_test"
	group  = "test_group"
)

func main() {
	brokers, _ := os.LookupEnv("KAFKA_BROKERS")
	input_topic, _ := os.LookupEnv("KAFKA_TOPIC_IN")
	output_topic, _ := os.LookupEnv("KAFKA_TOPIC_OUT")

	log.Printf("Got brokers: %v\n", brokers)
	log.Printf("Got input topic: %v\n", input_topic)
	log.Printf("Got output topic: %v\n", output_topic)

	consumer, err := InitNativeKafkaConsumer(client, brokers, group)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to start kafka consumer: %w", err))
	}

	defer consumer.Close()

	meta, err := consumer.GetMetadata(nil, true, 1000)
	log.Printf("Metadata: %#+v, %v", meta, err)

	producer, err := InitNativeKafkaProducer(client, brokers, "1", 30000)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to start kafka producer: %w", err))
	}

	defer producer.Close()

	err = consumer.SubscribeTopics([]string{input_topic}, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to subscribe to kafka topic: %w", err))
	}

	StartConsuming(context.TODO(), consumer, producer, output_topic)

	fmt.Print("Finished\n")
}

func StartConsuming(ctx context.Context, consumer *kafka.Consumer, producer *kafka.Producer, outTopic string) {
	log.Println("start consuming events")

	run := true

	for run {
		msg, err := consumer.ReadMessage(time.Second * 1)
		if err != nil {
			kErr, ok := err.(kafka.Error)
			if ok && kErr.IsTimeout() {
				log.Println(fmt.Errorf("read timeout: %w", kErr))
				continue
			}

			log.Println(fmt.Errorf("failed to read message: %w", err))
			continue
		}

		log.Printf("got message: %s\n", string(msg.Value))

		outputText := string(msg.Value) + "-from-internal"
		output := MakeMsg(outTopic, string(msg.Key), outputText)

		err = producer.Produce(&output, nil)

		if err != nil {
			log.Println(fmt.Errorf("failed to write message: %w", err))
			continue
		}

		log.Printf("written: %s\n", outputText)
	}
}

func MakeMsg(topic, key string, message interface{}) kafka.Message {
	headers := []kafka.Header{
		{
			Key: "DateAdd", Value: []byte(time.Now().Format(time.RFC3339Nano)),
		},
		{
			Key: "MessageId", Value: []byte(uuid.NewString()),
		},
	}

	messageJson, _ := json.Marshal(message)

	keyJson, _ := json.Marshal(key)

	msg := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value:   messageJson,
		Key:     keyJson,
		Headers: headers,
	}

	return msg
}

func InitNativeKafkaProducer(
	clientID string,
	brokers string,
	acks string,
	bufMaxMsg int,
) (*kafka.Producer, error) {
	cfg := kafka.ConfigMap{
		"bootstrap.servers":            brokers,
		"client.id":                    clientID,
		"acks":                         acks,
		"queue.buffering.max.messages": bufMaxMsg,
		"go.delivery.reports":          false,
	}

	p, err := kafka.NewProducer(&cfg)
	if err != nil {
		slog.Error("new producer", err)
		return nil, err
	}

	slog.Info(fmt.Sprintf("kafka producer %s created", clientID))

	return p, nil
}

func InitNativeKafkaConsumer(
	clientID string,
	brokers string,
	group string,
) (*kafka.Consumer, error) {
	config := kafka.ConfigMap{
		"bootstrap.servers":       brokers,
		"group.id":                group,
		"client.id":               clientID,
		"auto.offset.reset":       "earliest",
		"auto.commit.interval.ms": 3000,
	}

	c, err := kafka.NewConsumer(&config)
	if err != nil {
		return nil, errors.Wrapf(err, "create kafka consumer")
	}

	slog.Info(fmt.Sprintf("kafka consumer %s created", clientID))

	return c, nil
}
