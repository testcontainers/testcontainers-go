package canned

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/go-redis/redis/v7"
)

const (
	redisImage      = "redis"
	redisDefaultTag = "5.0"
	redisPort       = "6379/tcp"
)

// RedisContainerRequest completes GenericContainerRequest
type RedisContainerRequest struct {
	testcontainers.GenericContainerRequest
}

// RedisContainer should always be created via NewRedisContainer
type RedisContainer struct {
	Container testcontainers.Container
	client    *redis.Client
	req       RedisContainerRequest
}

// GetClient returns a sql.DB connecting to the previously started Postgres DB.
// All the parameters are taken from the previous RedisContainerRequest
func (c *RedisContainer) GetClient(ctx context.Context) (*redis.Client, error) {

	host, err := c.Container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := c.Container.MappedPort(ctx, redisPort)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, mappedPort.Int()),
		DB:   0,
	})

	return client, nil
}

// NewRedisContainer creates and (optionally) starts a Redis instance.
func NewRedisContainer(ctx context.Context, req RedisContainerRequest) (*RedisContainer, error) {

	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}

	// With the current logic it's not really possible to allow other ports...
	req.ExposedPorts = []string{redisPort}

	if req.Env == nil {
		req.Env = map[string]string{}
	}

	// Set the default values if none were provided in the request
	if req.Image == "" && req.FromDockerfile.Context == "" {
		req.Image = fmt.Sprintf("%s:%s", redisImage, redisDefaultTag)
	}

	req.WaitingFor = wait.ForListeningPort(redisPort)

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	redisC := &RedisContainer{
		Container: c,
		req:       req,
	}

	if req.Started {
		if err := c.Start(ctx); err != nil {
			return redisC, errors.Wrap(err, "failed to start container")
		}
	}

	return redisC, nil
}
