package examples_test

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func Example_withTLS() {
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

	// connectionString {
	uri, err := redisContainer.ConnectionString(ctx)
	// }
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

	// tlsConfig {
	options.TLSConfig = redisContainer.TLSConfig()
	// }

	client := redis.NewClient(options)
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close redis client: %s", err)
		}
	}()
	defer func(ctx context.Context, client *redis.Client) {
		err := client.FlushAll(ctx).Err()
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
