package testcontainers

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
)

type TestcontainersClient struct {
	*client.Client // client is embedded into our own client
}

func NewDockerClient() (cli *client.Client, err error) {
	client, err := testcontainersdocker.NewClient(context.Background())
	if err != nil {
		return nil, err
	}

	tcClient := &TestcontainersClient{
		Client: client,
	}

	return tcClient.Client, nil
}
