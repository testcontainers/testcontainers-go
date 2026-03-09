package memcached_test

import (
	"context"
	"fmt"
	"log"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/memcached"
)

func ExampleRun() {
	// runMemcachedContainer {
	ctx := context.Background()

	memcachedContainer, err := memcached.Run(ctx, "memcached:1.6-alpine")
	defer func() {
		if err := testcontainers.TerminateContainer(memcachedContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := memcachedContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connect() {
	ctx := context.Background()

	memcachedContainer, err := memcached.Run(ctx, "memcached:1.6-alpine")
	defer func() {
		if err := testcontainers.TerminateContainer(memcachedContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// hostPort {
	hostPort, err := memcachedContainer.HostPort(ctx)
	if err != nil {
		log.Printf("failed to get host and port: %s", err)
		return
	}
	// }

	mc := memcache.New(hostPort)

	err = mc.Set(&memcache.Item{Key: "foo", Value: []byte("my value")})
	if err != nil {
		log.Printf("failed to set item: %s", err)
		return
	}

	it, err := mc.Get("foo")
	if err != nil {
		log.Printf("failed to get item: %s", err)
		return
	}

	fmt.Println(string(it.Value))

	// Output:
	// my value
}
