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
	ctr, err := testcontainers.Run(ctx, "nginx",
		testcontainers.WithExposedPorts("80/tcp"),
		testcontainers.WithReuseByName("nginx"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/").WithStartupTimeout(10*time.Second)),
	)
	var nginxC *nginxContainer
	if ctr != nil {
		nginxC = &nginxContainer{Container: ctr}
	}
	if err != nil {
		return nginxC, err
	}

	ip, err := ctr.Host(ctx)
	if err != nil {
		return nginxC, err
	}

	mappedPort, err := ctr.MappedPort(ctx, "80")
	if err != nil {
		return nginxC, err
	}

	nginxC.URI = fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())
	return nginxC, nil
}
