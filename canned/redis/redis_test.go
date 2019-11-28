package redis

import (
	"context"
	"github.com/go-redis/redis"
	"gotest.tools/assert"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

func TestSetInRedis(t *testing.T) {
	ctx := context.Background()

	c, err := NewContainer(ctx, ContainerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Container.Terminate(ctx)

	addr, err := c.ConnectURL(ctx)
	if err != nil {
		t.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
	defer client.Close()

	err = client.Set("key", "value", 0).Err()
	if err != nil {
		t.Fatal(err)
	}

	got := client.Get("key")
	if got.Err() != nil {
		t.Fatal(got.Err())
	}

	assert.Equal(t, got.Val(), "value")
}

func ExampleNewContainer() {
	ctx := context.Background()

	c, _ := NewContainer(ctx, ContainerRequest{
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			Started: true,
		},
	})

	defer c.Container.Terminate(ctx)
}
