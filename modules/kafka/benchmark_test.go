package kafka_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func startStopBenchmark(b *testing.B, image string) {
	for b.Loop() {
		kafkaContainer, err := kafka.Run(context.Background(),
			image,
		)
		if err != nil {
			b.Fatalf("failed to start container: %s", err)
		}

		if err := testcontainers.TerminateContainer(kafkaContainer); err != nil {
			b.Errorf("failed to terminate container: %s", err)
		}
	}
}

func BenchmarkConfluentStartStop(b *testing.B) {
	startStopBenchmark(b, "confluentinc/confluent-local:7.5.0")
}

func BenchmarkApacheNativeStartStop(b *testing.B) {
	startStopBenchmark(b, "apache/kafka-native:4.0.1")
}

func BenchmarkApacheStartStop(b *testing.B) {
	startStopBenchmark(b, "apache/kafka:4.0.1")
}
