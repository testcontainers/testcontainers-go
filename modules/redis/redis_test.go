package redis_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestIntegrationSetGet(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx, "redis:7")
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	assertSetsGets(t, ctx, redisContainer, 1)
}

func TestRedisWithConfigFile(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx, "redis:7", tcredis.WithConfigFile(filepath.Join("testdata", "redis7.conf")))
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	jsonInspect, err := redisContainer.Inspect(ctx)
	require.NoError(t, err)

	require.Contains(t, jsonInspect.Config.Cmd, "/usr/local/redis.conf")

	assertSetsGets(t, ctx, redisContainer, 1)
}

func TestRedisWithLogLevel(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx, "redis:7", tcredis.WithLogLevel(tcredis.LogLevelVerbose))
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	jsonInspect, err := redisContainer.Inspect(ctx)
	require.NoError(t, err)

	require.Contains(t, jsonInspect.Config.Cmd, "--loglevel", "verbose")

	assertSetsGets(t, ctx, redisContainer, 10)
}

func TestRedisWithSnapshotting(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx, "redis:7", tcredis.WithSnapshotting(10, 1))
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	jsonInspect, err := redisContainer.Inspect(ctx)
	require.NoError(t, err)

	require.Contains(t, jsonInspect.Config.Cmd, "--save", "10", "1")

	assertSetsGets(t, ctx, redisContainer, 10)
}

func TestRedisWithTLS(t *testing.T) {
	ctx := context.Background()

	t.Run("mtls-disabled", func(t *testing.T) {
		redisContainer, err := tcredis.Run(ctx, "redis:7", tcredis.WithTLS())
		testcontainers.CleanupContainer(t, redisContainer)
		require.NoError(t, err)

		assertSetsGets(t, ctx, redisContainer, 1)
	})

	t.Run("mtls-enabled", func(t *testing.T) {
		redisContainer, err := tcredis.Run(ctx, "redis:7", tcredis.WithTLS(), testcontainers.WithCmdArgs("--tls-auth-clients", "no"))
		testcontainers.CleanupContainer(t, redisContainer)
		require.NoError(t, err)

		assertSetsGets(t, ctx, redisContainer, 1)
	})
}

func assertSetsGets(t *testing.T, ctx context.Context, redisContainer *tcredis.RedisContainer, keyCount int) {
	t.Helper()

	// connectionString {
	uri, err := redisContainer.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	// You will likely want to wrap your Redis package of choice in an
	// interface to aid in unit testing and limit lock-in throughout your
	// codebase but that's out of scope for this example
	options, err := redis.ParseURL(uri)
	require.NoError(t, err)

	// tlsConfig {
	options.TLSConfig = redisContainer.TLSConfig()
	// }

	client := redis.NewClient(options)
	defer func(t *testing.T, ctx context.Context, client *redis.Client) {
		t.Helper()
		require.NoError(t, flushRedis(ctx, *client))
	}(t, ctx, client)

	t.Log("pinging redis")
	pong, err := client.Ping(ctx).Result()
	require.NoError(t, err)

	t.Log("received response from redis")

	require.Equalf(t, "PONG", pong, "received unexpected response from redis: %s", pong)

	for i := 0; i < keyCount; i++ {
		// Set data
		key := fmt.Sprintf("{user.%s}.favoritefood.%d", uuid.NewString(), i)
		value := fmt.Sprintf("Cabbage Biscuits %d", i)

		ttl, _ := time.ParseDuration("2h")
		err = client.Set(ctx, key, value, ttl).Err()
		require.NoError(t, err)

		// Get data
		savedValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)

		require.Equal(t, savedValue, value)
	}
}

func flushRedis(ctx context.Context, client redis.Client) error {
	return client.FlushAll(ctx).Err()
}
