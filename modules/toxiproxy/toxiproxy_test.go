package toxiproxy_test

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"testing"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	tctoxiproxy "github.com/testcontainers/testcontainers-go/modules/toxiproxy"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	ctr, err := tctoxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}

//go:embed testdata/toxiproxy.json
var configFile []byte

func TestRun_withConfigFile(t *testing.T) {
	ctx := context.Background()

	nw, err := network.New(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, nw.Remove(ctx))
	})

	redisContainer, err := tcredis.Run(
		ctx,
		"redis:6-alpine",
		network.WithNetwork([]string{"redis"}, nw),
	)
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	toxiproxyContainer, err := tctoxiproxy.Run(
		ctx,
		"ghcr.io/shopify/toxiproxy:2.12.0",
		tctoxiproxy.WithConfigFile(bytes.NewReader(configFile)),
		network.WithNetwork([]string{"toxiproxy"}, nw),
	)
	testcontainers.CleanupContainer(t, toxiproxyContainer)
	require.NoError(t, err)

	toxiURI, err := toxiproxyContainer.URI(ctx)
	require.NoError(t, err)

	toxiproxyClient := toxiproxy.NewClient(toxiURI)

	toxiproxyProxyPort, err := toxiproxyContainer.MappedPort(ctx, "8666/tcp")
	require.NoError(t, err)

	toxiproxyProxyHostIP, err := toxiproxyContainer.Host(ctx)
	require.NoError(t, err)

	// Create a redis client that connects to the toxiproxy container.
	// We are defining a read timeout of 2 seconds, because we are adding
	// a latency toxic of 1 second to the request, +/- 100ms jitter.
	redisURI := fmt.Sprintf("redis://%s:%s?read_timeout=2s", toxiproxyProxyHostIP, toxiproxyProxyPort.Port())

	options, err := redis.ParseURL(redisURI)
	require.NoError(t, err)

	redisCli := redis.NewClient(options)
	t.Cleanup(func() {
		require.NoError(t, redisCli.FlushAll(ctx).Err())
	})

	key := fmt.Sprintf("{user.%s}.favoritefood", uuid.NewString())
	value := "Cabbage Biscuits"
	ttl, err := time.ParseDuration("2h")
	require.NoError(t, err)

	err = redisCli.Set(ctx, key, value, ttl).Err()
	require.NoError(t, err)

	const (
		latency = 1_000
		jitter  = 200
	)

	// Add a latency toxic to the proxy
	toxicOptions := &toxiproxy.ToxicOptions{
		ProxyName: "redis", // name of the proxy in the config file
		ToxicName: "latency_down",
		ToxicType: "latency",
		Toxicity:  1.0,
		Stream:    "downstream",
		Attributes: map[string]any{
			"latency": latency,
			"jitter":  jitter,
		},
	}
	_, err = toxiproxyClient.AddToxic(toxicOptions)
	require.NoError(t, err)

	start := time.Now()
	// Get data
	savedValue, err := redisCli.Get(ctx, key).Result()
	require.NoError(t, err)

	duration := time.Since(start)

	t.Logf("Duration: %s\n", duration)

	// The value is retrieved successfully
	require.Equal(t, value, savedValue)

	// Check that latency is within expected range (900ms-1100ms)
	// The latency toxic adds 1000ms (1000ms +/- 100ms jitter)
	minDuration := (latency - jitter) * time.Millisecond
	maxDuration := (latency + jitter) * time.Millisecond
	require.True(t, duration >= minDuration && duration <= maxDuration)
}
