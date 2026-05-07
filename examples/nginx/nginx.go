package nginx

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type nginxContainer struct {
	testcontainers.Container
	URI string
}

func startContainer(ctx context.Context) (*nginxContainer, error) {
	ctr, err := testcontainers.Run(
		ctx, "nginx",
		testcontainers.WithExposedPorts("80/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/").WithStartupTimeout(10*time.Second)),
	)
	var nginxC *nginxContainer
	if ctr != nil {
		nginxC = &nginxContainer{Container: ctr}
	}
	if err != nil {
		return nginxC, err
	}

	endpoint, err := ctr.PortEndpoint(ctx, "80", "http")
	if err != nil {
		return nginxC, err
	}

	nginxC.URI = endpoint
	return nginxC, nil
}
