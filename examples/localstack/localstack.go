package localstack

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// localStackContainer represents the LocalStack container type used in the module
type localStackContainer struct {
	testcontainers.Container
}

// setupLocalStack creates an instance of the LocalStack container type
func setupLocalStack(ctx context.Context) (*localStackContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "localstack/localstack:0.11.2",
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &localStackContainer{Container: container}, nil
}
