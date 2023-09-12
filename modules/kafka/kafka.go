package kafka

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// KafkaContainer represents the Kafka container type used in the module
type KafkaContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Kafka container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "confluentinc/cp-kafka:7.3.3",
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &KafkaContainer{Container: container}, nil
}
