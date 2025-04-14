package toxiproxy

import (
	"context"
	"fmt"
	"testing"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestToxiproxy(t *testing.T) {
	ctx := context.Background()

	newNetwork, err := network.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, newNetwork)

	networkName := newNetwork.Name

	toxiproxyContainer, err := startContainer(ctx, networkName, []string{"toxiproxy"})
	testcontainers.CleanupContainer(t, toxiproxyContainer)
	require.NoError(t, err)

	redisContainer, err := setupRedis(ctx, networkName, []string{"redis"})
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	toxiproxyClient := toxiproxy.NewClient(toxiproxyContainer.URI)
	proxy, err := toxiproxyClient.CreateProxy("redis", "0.0.0.0:8666", "redis:6379")
	require.NoError(t, err)

	toxiproxyProxyPort, err := toxiproxyContainer.MappedPort(ctx, "8666")
	require.NoError(t, err)

	toxiproxyProxyHostIP, err := toxiproxyContainer.Host(ctx)
	require.NoError(t, err)

	redisURI := fmt.Sprintf("redis://%s:%s?read_timeout=2s", toxiproxyProxyHostIP, toxiproxyProxyPort.Port())

	options, err := redis.ParseURL(redisURI)
	require.NoError(t, err)
	redisClient := redis.NewClient(options)

	defer func() {
		require.NoError(t, flushRedis(ctx, *redisClient))
	}()

	// Set data
	key := fmt.Sprintf("{user.%s}.favoritefood", uuid.NewString())
	value := "Cabbage Biscuits"
	ttl, _ := time.ParseDuration("2h")
	err = redisClient.Set(ctx, key, value, ttl).Err()
	require.NoError(t, err)

	_, err = proxy.AddToxic("latency_down", "latency", "downstream", 1.0, toxiproxy.Attributes{
		"latency": 1000,
		"jitter":  100,
	})
	require.NoError(t, err)

	// Get data
	savedValue, err := redisClient.Get(ctx, key).Result()
	require.NoError(t, err)
	require.Equal(t, value, savedValue)
}

func flushRedis(ctx context.Context, client redis.Client) error {
	return client.FlushAll(ctx).Err()
}
