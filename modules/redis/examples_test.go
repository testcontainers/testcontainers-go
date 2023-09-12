package redis_test

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func ExampleRunContainer() {
	// runRedisContainer {
	ctx := context.Background()

	redisContainer, err := redis.RunContainer(ctx,
		testcontainers.WithImage("docker.io/redis:7"),
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
		redis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := redisContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
