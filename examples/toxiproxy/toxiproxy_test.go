package toxiproxy

import (
	"context"
	"fmt"
	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"testing"
	"time"
)

func TestToxiproxy(t *testing.T) {
	ctx := context.Background()

	newNetwork, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		ProviderType: testcontainers.ProviderDocker,
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           "newNetwork",
			CheckDuplicate: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	toxiproxyContainer, err := startContainer(ctx, "newNetwork", []string{"toxiproxy"})
	if err != nil {
		t.Fatal(err)
	}

	redisContainer, err := setupRedis(ctx, "newNetwork", []string{"redis"})
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := toxiproxyContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
		if err := newNetwork.Remove(ctx); err != nil {
			t.Fatalf("failed to terminate network: %s", err)
		}
	})

	toxiproxyClient := toxiproxy.NewClient(toxiproxyContainer.URI)
	proxy, err := toxiproxyClient.CreateProxy("redis", "0.0.0.0:8666", "redis:6379")
	if err != nil {
		t.Fatal(err)
	}

	toxiproxyProxyPort, err := toxiproxyContainer.MappedPort(ctx, "8666")
	if err != nil {
		t.Fatal(err)
	}

	toxiproxyProxyHostIP, err := toxiproxyContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	redisUri := fmt.Sprintf("redis://%s:%s?read_timeout=2s", toxiproxyProxyHostIP, toxiproxyProxyPort.Port())

	options, err := redis.ParseURL(redisUri)
	if err != nil {
		t.Fatal(err)
	}
	redisClient := redis.NewClient(options)
	defer flushRedis(ctx, *redisClient)

	// Set data
	key := fmt.Sprintf("{user.%s}.favoritefood", uuid.NewString())
	value := "Cabbage Biscuits"
	ttl, _ := time.ParseDuration("2h")
	err = redisClient.Set(ctx, key, value, ttl).Err()
	if err != nil {
		t.Fatal(err)
	}

	_, err = proxy.AddToxic("latency_down", "latency", "downstream", 1.0, toxiproxy.Attributes{
		"latency": 1000,
		"jitter":  100,
	})
	if err != nil {
		return
	}

	// Get data
	savedValue, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}

	// perform assertions
	if savedValue != value {
		t.Fatalf("Expected value %s. Got %s.", savedValue, value)
	}
}

func flushRedis(ctx context.Context, client redis.Client) error {
	return client.FlushAll(ctx).Err()
}
