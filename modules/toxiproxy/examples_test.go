package toxiproxy_test

import (
	"context"
	"fmt"
	"log"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	tctoxiproxy "github.com/testcontainers/testcontainers-go/modules/toxiproxy"
	"github.com/testcontainers/testcontainers-go/network"
)

func ExampleRun() {
	// runToxiproxyContainer {
	ctx := context.Background()

	toxiproxyContainer, err := tctoxiproxy.Run(
		ctx,
		"ghcr.io/shopify/toxiproxy:2.12.0",
	)
	defer func() {
		if err := testcontainers.TerminateContainer(toxiproxyContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := toxiproxyContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_addLatency() {
	ctx := context.Background()

	nw, err := network.New(ctx)
	if err != nil {
		log.Printf("failed to create network: %v", err)
		return
	}
	defer func() {
		if err := nw.Remove(ctx); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()

	redisContainer, err := tcredis.Run(
		ctx,
		"redis:6-alpine",
		network.WithNetwork([]string{"redis"}, nw),
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

	// defineContainerExposingPort {
	const proxyPort = "8666"

	// No need to create a proxy, as we are programmatically adding it below.
	toxiproxyContainer, err := tctoxiproxy.Run(
		ctx,
		"ghcr.io/shopify/toxiproxy:2.12.0",
		network.WithNetwork([]string{"toxiproxy"}, nw),
		// explicitly expose the ports that will be proxied using the programmatic API
		// of the toxiproxy client. Otherwise, the ports will not be exposed and the
		// toxiproxy client will not be able to connect to the proxy.
		testcontainers.WithExposedPorts(proxyPort+"/tcp"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(toxiproxyContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	// createToxiproxyClient {
	toxiURI, err := toxiproxyContainer.URI(ctx)
	if err != nil {
		log.Printf("failed to get toxiproxy container uri: %s", err)
		return
	}

	toxiproxyClient := toxiproxy.NewClient(toxiURI)
	// }

	// createProxy {
	// Create the proxy using the network alias of the redis container,
	// as they run on the same network.
	proxy, err := toxiproxyClient.CreateProxy("redis", "0.0.0.0:"+proxyPort, "redis:6379")
	if err != nil {
		log.Printf("failed to create proxy: %s", err)
		return
	}
	// }

	toxiproxyProxyPort, err := toxiproxyContainer.MappedPort(ctx, proxyPort+"/tcp")
	if err != nil {
		log.Printf("failed to get toxiproxy container port: %s", err)
		return
	}

	toxiproxyProxyHostIP, err := toxiproxyContainer.Host(ctx)
	if err != nil {
		log.Printf("failed to get toxiproxy container host: %s", err)
		return
	}

	// createRedisClient {
	// Create a redis client that connects to the toxiproxy container.
	// We are defining a read timeout of 2 seconds, because we are adding
	// a latency toxic of 1 second to the request, +/- 100ms jitter.
	redisURI := fmt.Sprintf("redis://%s:%s?read_timeout=2s", toxiproxyProxyHostIP, toxiproxyProxyPort.Port())

	options, err := redis.ParseURL(redisURI)
	if err != nil {
		log.Printf("failed to parse url: %s", err)
		return
	}

	redisCli := redis.NewClient(options)
	defer func() {
		if err := redisCli.FlushAll(ctx).Err(); err != nil {
			log.Printf("failed to flush redis: %s", err)
		}
	}()
	// }

	key := fmt.Sprintf("{user.%s}.favoritefood", uuid.NewString())
	value := "Cabbage Biscuits"
	ttl, err := time.ParseDuration("2h")
	if err != nil {
		log.Printf("failed to parse duration: %s", err)
		return
	}

	err = redisCli.Set(ctx, key, value, ttl).Err()
	if err != nil {
		log.Printf("failed to set data: %s", err)
		return
	}

	// addLatencyToxic {
	const (
		latency = 1_000
		jitter  = 200
	)
	// Add a latency toxic to the proxy
	_, err = proxy.AddToxic("latency_down", "latency", "downstream", 1.0, toxiproxy.Attributes{
		"latency": latency,
		"jitter":  jitter,
	})
	if err != nil {
		log.Printf("failed to add toxic: %s", err)
		return
	}
	// }

	start := time.Now()
	// Get data
	savedValue, err := redisCli.Get(ctx, key).Result()
	if err != nil {
		log.Printf("failed to get data: %s", err)
		return
	}
	duration := time.Since(start)

	log.Println("Duration:", duration)

	// The value is retrieved successfully
	fmt.Println(savedValue)

	// Check that latency is within expected range (200ms-1200ms)
	// The latency toxic adds 1000ms (1000ms +/- 200ms jitter)
	minDuration := (latency - jitter) * time.Millisecond
	maxDuration := (latency + jitter) * time.Millisecond
	fmt.Printf("Duration is between %dms and %dms: %v\n",
		minDuration.Milliseconds(),
		maxDuration.Milliseconds(),
		duration >= minDuration && duration <= maxDuration)

	// Output:
	// Cabbage Biscuits
	// Duration is between 800ms and 1200ms: true
}

func ExampleRun_connectionCut() {
	ctx := context.Background()

	nw, err := network.New(ctx)
	if err != nil {
		log.Printf("failed to create network: %v", err)
		return
	}
	defer func() {
		if err := nw.Remove(ctx); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()

	redisContainer, err := tcredis.Run(
		ctx,
		"redis:6-alpine",
		network.WithNetwork([]string{"redis"}, nw),
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

	toxiproxyContainer, err := tctoxiproxy.Run(
		ctx,
		"ghcr.io/shopify/toxiproxy:2.12.0",
		// We create a proxy named "redis" that points to the redis container.
		tctoxiproxy.WithProxy("redis", "redis:6379"),
		network.WithNetwork([]string{"toxiproxy"}, nw),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(toxiproxyContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// getProxiedEndpoint {
	proxiedRedisHost, proxiedRedisPort, err := toxiproxyContainer.ProxiedEndpoint(8666)
	if err != nil {
		log.Printf("failed to get toxiproxy container port: %s", err)
		return
	}
	// }

	toxiURI, err := toxiproxyContainer.URI(ctx)
	if err != nil {
		log.Printf("failed to get toxiproxy container uri: %s", err)
		return
	}

	toxiproxyClient := toxiproxy.NewClient(toxiURI)

	// Retrieve the existing proxy
	proxies, err := toxiproxyClient.Proxies()
	if err != nil {
		log.Printf("failed to get proxies: %s", err)
		return
	}

	proxy := proxies["redis"]

	// readProxiedEndpoint {
	// Create a redis client that connects to the toxiproxy container.
	// We are defining a read timeout of 2 seconds, because we are adding
	// a latency toxic of 1.1 seconds to the request.
	redisURI := fmt.Sprintf("redis://%s:%s?read_timeout=2s", proxiedRedisHost, proxiedRedisPort)
	// }

	options, err := redis.ParseURL(redisURI)
	if err != nil {
		log.Printf("failed to parse url: %s", err)
		return
	}

	redisCli := redis.NewClient(options)
	defer func() {
		if err := redisCli.FlushAll(ctx).Err(); err != nil {
			log.Printf("failed to flush redis: %s", err)
		}
	}()

	key := fmt.Sprintf("{user.%s}.favoritefood", uuid.NewString())
	value := "Cabbage Biscuits"
	ttl, err := time.ParseDuration("2h")
	if err != nil {
		log.Printf("failed to parse duration: %s", err)
		return
	}

	err = redisCli.Set(ctx, key, value, ttl).Err()
	if err != nil {
		log.Printf("failed to set data: %s", err)
		return
	}

	// Disable the proxy
	err = proxy.Disable()
	if err != nil {
		log.Printf("failed to disable proxy: %s", err)
		return
	}

	// Get data
	savedValue, err := redisCli.Get(ctx, key).Result()
	if err == nil {
		log.Printf("proxy is disabled, but we got data")
		return
	}

	// The value is not retrieved at all, so it's empty
	fmt.Println(savedValue)

	// Re-enable the proxy
	err = proxy.Enable()
	if err != nil {
		log.Printf("failed to enable proxy: %s", err)
		return
	}

	savedValue, err = redisCli.Get(ctx, key).Result()
	if err != nil {
		log.Printf("failed to get data: %s", err)
		return
	}

	// The value is retrieved successfully
	fmt.Println(savedValue)

	// Output:
	//
	// Cabbage Biscuits
}
