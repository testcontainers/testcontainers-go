package nginx

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type nginxContainer struct {
	testcontainers.Container
	URI string
}

func startContainer(ctx context.Context) (*nginxContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	var nginxC *nginxContainer
	if container != nil {
		nginxC = &nginxContainer{Container: container}
	}
	if err != nil {
		return nginxC, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nginxC, err
	}

	mappedPort, err := container.MappedPort(ctx, "80")
	if err != nil {
		return nginxC, err
	}

	nginxC.URI = fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())
	return nginxC, nil
}
