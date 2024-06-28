package toxiproxy

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type redisContainer struct {
	*testcontainers.DockerContainer
}

func setupRedis(ctx context.Context, network string, networkAlias []string) (*redisContainer, error) {
	req := testcontainers.Request{
		Image:        "redis:6",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("* Ready to accept connections"),
		Networks: []string{
			network,
		},
		NetworkAliases: map[string][]string{
			network: networkAlias,
		},
		Started: true,
	}
	ctr, err := testcontainers.Run(ctx, req)
	if err != nil {
		return nil, err
	}

	return &redisContainer{DockerContainer: ctr}, nil
}
