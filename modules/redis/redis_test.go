package redis

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestIntegrationSetGet(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := RunContainer(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertSetsGets(t, ctx, redisContainer, 1)
}

func TestRedisWithConfigFile(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := RunContainer(ctx, WithConfigFile(filepath.Join("testdata", "redis7.conf")))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertSetsGets(t, ctx, redisContainer, 1)
}

func TestRedisWithImage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		image string
	}{
		{
			name:  "Redis6",
			image: "docker.io/redis:6",
		},
		{
			name:  "Redis7",
			image: "docker.io/redis:7",
		},
		{
			name: "Redis Stack",
			// redisStackImage {
			image: "docker.io/redis/redis-stack:latest",
			// }
		},
		{
			name: "Redis Stack Server",
			// redisStackServerImage {
			image: "docker.io/redis/redis-stack-server:latest",
			// }
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisContainer, err := RunContainer(ctx, testcontainers.WithImage(tt.image), WithConfigFile(filepath.Join("testdata", "redis6.conf")))
			require.NoError(t, err)
			t.Cleanup(func() {
				if err := redisContainer.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})

			assertSetsGets(t, ctx, redisContainer, 1)
		})
	}
}

func TestRedisWithLogLevel(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := RunContainer(ctx, WithLogLevel(LogLevelVerbose))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertSetsGets(t, ctx, redisContainer, 10)
}

func TestRedisWithSnapshotting(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := RunContainer(ctx, WithSnapshotting(10, 1))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertSetsGets(t, ctx, redisContainer, 10)
}

func assertSetsGets(t *testing.T, ctx context.Context, redisContainer *RedisContainer, keyCount int) {
	// connectionString {
	uri, err := redisContainer.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	// You will likely want to wrap your Redis package of choice in an
	// interface to aid in unit testing and limit lock-in throughout your
	// codebase but that's out of scope for this example
	options, err := redis.ParseURL(uri)
	require.NoError(t, err)

	client := redis.NewClient(options)
	defer func(t *testing.T, ctx context.Context, client *redis.Client) {
		require.NoError(t, flushRedis(ctx, *client))
	}(t, ctx, client)

	t.Log("pinging redis")
	pong, err := client.Ping(ctx).Result()
	require.NoError(t, err)

	t.Log("received response from redis")

	if pong != "PONG" {
		t.Fatalf("received unexpected response from redis: %s", pong)
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

func flushRedis(ctx context.Context, client redis.Client) error {
	return client.FlushAll(ctx).Err()
}

func TestWithConfigFile(t *testing.T) {
	tests := []struct {
		name         string
		cmds         []string
		expectedCmds []string
	}{
		{
			name:         "no existing command",
			cmds:         []string{},
			expectedCmds: []string{redisServerProcess, "/usr/local/redis.conf"},
		},
		{
			name:         "existing redis-server command as first argument",
			cmds:         []string{redisServerProcess, "a", "b", "c"},
			expectedCmds: []string{redisServerProcess, "/usr/local/redis.conf", "a", "b", "c"},
		},
		{
			name:         "non existing redis-server command",
			cmds:         []string{"a", "b", "c"},
			expectedCmds: []string{redisServerProcess, "/usr/local/redis.conf", "a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Cmd: tt.cmds,
				},
			}

			WithConfigFile("redis.conf")(req)

			require.Equal(t, tt.expectedCmds, req.Cmd)
		})
	}
}

func TestWithLogLevel(t *testing.T) {
	tests := []struct {
		name         string
		cmds         []string
		expectedCmds []string
	}{
		{
			name:         "no existing command",
			cmds:         []string{},
			expectedCmds: []string{redisServerProcess, "--loglevel", "debug"},
		},
		{
			name:         "existing redis-server command as first argument",
			cmds:         []string{redisServerProcess, "a", "b", "c"},
			expectedCmds: []string{redisServerProcess, "a", "b", "c", "--loglevel", "debug"},
		},
		{
			name:         "non existing redis-server command",
			cmds:         []string{"a", "b", "c"},
			expectedCmds: []string{redisServerProcess, "a", "b", "c", "--loglevel", "debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Cmd: tt.cmds,
				},
			}

			WithLogLevel(LogLevelDebug)(req)

			require.Equal(t, tt.expectedCmds, req.Cmd)
		})
	}
}

func TestWithSnapshotting(t *testing.T) {
	tests := []struct {
		name         string
		cmds         []string
		expectedCmds []string
		seconds      int
		changedKeys  int
	}{
		{
			name:         "no existing command",
			cmds:         []string{},
			seconds:      60,
			changedKeys:  100,
			expectedCmds: []string{redisServerProcess, "--save", "60", "100"},
		},
		{
			name:         "existing redis-server command as first argument",
			cmds:         []string{redisServerProcess, "a", "b", "c"},
			seconds:      60,
			changedKeys:  100,
			expectedCmds: []string{redisServerProcess, "a", "b", "c", "--save", "60", "100"},
		},
		{
			name:         "non existing redis-server command",
			cmds:         []string{"a", "b", "c"},
			seconds:      60,
			changedKeys:  100,
			expectedCmds: []string{redisServerProcess, "a", "b", "c", "--save", "60", "100"},
		},
		{
			name:         "existing redis-server command as first argument",
			cmds:         []string{redisServerProcess, "a", "b", "c"},
			seconds:      0,
			changedKeys:  0,
			expectedCmds: []string{redisServerProcess, "a", "b", "c", "--save", "1", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Cmd: tt.cmds,
				},
			}

			WithSnapshotting(tt.seconds, tt.changedKeys)(req)

			require.Equal(t, tt.expectedCmds, req.Cmd)
		})
	}
}
