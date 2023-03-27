package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSetGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// createRedisContainer {
	redisContainer, err := StartContainer(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
	// }

	// connectionString {
	uri, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)
	// }

	// You will likely want to wrap your Redis package of choice in an
	// interface to aid in unit testing and limit lock-in throughtout your
	// codebase but that's out of scope for this example
	options, err := redis.ParseURL(uri)
	require.NoError(t, err)

	client := redis.NewClient(options)
	defer flushRedis(ctx, *client)

	t.Log("pinging redis")
	pong, err := client.Ping(ctx).Result()
	require.NoError(t, err)

	t.Log("received response from redis")

	if pong != "PONG" {
		t.Fatalf("received unexpected response from redis: %s", pong)
	}

	// Set data
	key := fmt.Sprintf("{user.%s}.favoritefood", uuid.NewString())
	value := "Cabbage Biscuits"
	ttl, _ := time.ParseDuration("2h")
	err = client.Set(ctx, key, value, ttl).Err()
	require.NoError(t, err)

	// Get data
	savedValue, err := client.Get(ctx, key).Result()
	require.NoError(t, err)

	if savedValue != value {
		t.Fatalf("Expected value %s. Got %s.", savedValue, value)
	}
}

func flushRedis(ctx context.Context, client redis.Client) error {
	return client.FlushAll(ctx).Err()
}
