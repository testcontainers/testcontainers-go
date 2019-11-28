package redis

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	image = "redis"
	tag   = "5.0.7"
	port  = "6379/tcp"
)

// RedisContainerRequest completes GenericContainerRequest
type ContainerRequest struct {
	testcontainers.GenericContainerRequest
	Version string
}

// Container should always be created via NewContainer
type Container struct {
	Container testcontainers.Container
	req       ContainerRequest
}

func (c Container) ConnectURL(ctx context.Context) (string, error) {
	host, err := c.Container.Host(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get host")
	}

	mappedPort, err := c.Container.MappedPort(ctx, port)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get mapped port for %s", port)
	}

	return fmt.Sprintf("%s:%d", host, mappedPort.Int()), nil
}

// NewContainer creates and (optionally) starts a Redis instance.
func NewContainer(ctx context.Context, req ContainerRequest) (*Container, error) {

	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}

	// With the current logic it's not really possible to allow other ports...
	req.ExposedPorts = []string{port}

	if req.Env == nil {
		req.Env = map[string]string{}
	}

	if req.Version != "" && req.FromDockerfile.Context == "" {
		req.Image = fmt.Sprintf("%s:%s", image, req.Version)
	}

	if req.Image == "" && req.FromDockerfile.Context == "" {
		req.Image = fmt.Sprintf("%s:%s", image, tag)
	}

	req.WaitingFor = wait.NewHostPortStrategy(port)

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	res := &Container{
		Container: c,
		req:       req,
	}

	if err := c.Start(ctx); err != nil {
		return res, errors.Wrap(err, "failed to start container")
	}

	return res, nil
}
