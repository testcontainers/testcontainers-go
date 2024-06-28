package nginx

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type nginxContainer struct {
	*testcontainers.DockerContainer
	URI string
}

func startContainer(ctx context.Context) (*nginxContainer, error) {
	req := testcontainers.Request{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
		Started:      true,
	}
	ctr, err := testcontainers.Run(ctx, req)
	if err != nil {
		return nil, err
	}

	ip, err := ctr.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := ctr.MappedPort(ctx, "80")
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	return &nginxContainer{DockerContainer: ctr, URI: uri}, nil
}
