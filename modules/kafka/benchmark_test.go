package kafka_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func startStopBenchmark(b *testing.B, image string) {
	b.Helper()
	for b.Loop() {
		kafkaContainer, err := kafka.Run(context.Background(),
			image,
		)
		require.NoError(b, err)

		err = testcontainers.TerminateContainer(kafkaContainer)
		require.NoError(b, err)
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
