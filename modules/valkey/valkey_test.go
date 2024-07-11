package valkey_test

import (
	"context"
	"fmt"
	tcvalkey "github.com/testcontainers/testcontainers-go/modules/valkey"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSetGet(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "docker.io/valkey/valkey:7.2.5")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := valkeyContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertSetsGets(t, ctx, valkeyContainer, 1)
}

func TestValkeyWithConfigFile(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "docker.io/valkey/valkey:7.2.5", tcvalkey.WithConfigFile(filepath.Join("testdata", "valkey7.conf")))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := valkeyContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertSetsGets(t, ctx, valkeyContainer, 1)
}

func TestValkeyWithImage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		image string
	}{
		// There is only one release of Valkey at the time of writing
		{
			name:  "Valkey7.2.5",
			image: "docker.io/valkey/valkey:7.2.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valkeyContainer, err := tcvalkey.Run(ctx, tt.image, tcvalkey.WithConfigFile(filepath.Join("testdata", "valkey7.conf")))
			require.NoError(t, err)
			t.Cleanup(func() {
				if err := valkeyContainer.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})

			assertSetsGets(t, ctx, valkeyContainer, 1)
		})
	}
}

func TestValkeyWithLogLevel(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "docker.io/valkey/valkey:7.2.5", tcvalkey.WithLogLevel(tcvalkey.LogLevelVerbose))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := valkeyContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertSetsGets(t, ctx, valkeyContainer, 10)
}

func TestRedisWithSnapshotting(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "docker.io/valkey/valkey:7.2.5", tcvalkey.WithSnapshotting(10, 1))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := valkeyContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertSetsGets(t, ctx, valkeyContainer, 10)
}

func assertSetsGets(t *testing.T, ctx context.Context, valkeyContainer *tcvalkey.ValkeyContainer, keyCount int) {
	// connectionString {
	uri, err := valkeyContainer.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	// You will likely want to wrap your Valkey package of choice in an
	// interface to aid in unit testing and limit lock-in throughout your
	// codebase but that's out of scope for this example
	options, err := redis.ParseURL(uri)
	require.NoError(t, err)

	client := redis.NewClient(options)
	defer func(t *testing.T, ctx context.Context, client *redis.Client) {
		require.NoError(t, flushValkey(ctx, *client))
	}(t, ctx, client)

	t.Log("pinging valkey")
	pong, err := client.Ping(ctx).Result()
	require.NoError(t, err)

	t.Log("received response from valkey")

	if pong != "PONG" {
		t.Fatalf("received unexpected response from valkey: %s", pong)
	}

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

		if savedValue != value {
			t.Fatalf("Expected value %s. Got %s.", savedValue, value)
		}
	}
}

func flushValkey(ctx context.Context, client redis.Client) error {
	return client.FlushAll(ctx).Err()
}
