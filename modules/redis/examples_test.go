package redis_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/redis/go-redis/v9"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func ExampleRun() {
	// runRedisContainer {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx,
		"redis:7",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
		tcredis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := redisContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_withTLS() {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx,
		"redis:7",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
		tcredis.WithTLS(),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	if redisContainer.TLSConfig() == nil {
		log.Println("TLS is not enabled")
		return
	}

	uri, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	// You will likely want to wrap your Redis package of choice in an
	// interface to aid in unit testing and limit lock-in throughout your
	// codebase but that's out of scope for this example
	options, err := redis.ParseURL(uri)
	if err != nil {
		log.Printf("failed to parse connection string: %s", err)
		return
	}

	options.TLSConfig = redisContainer.TLSConfig()

	client := redis.NewClient(options)
	defer func(ctx context.Context, client *redis.Client) {
		err := flushRedis(ctx, *client)
		if err != nil {
			log.Printf("failed to flush redis: %s", err)
		}
	}(ctx, client)

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("failed to ping redis: %s", err)
		return
	}

	fmt.Println(pong)

	// Output:
	// PONG
}
