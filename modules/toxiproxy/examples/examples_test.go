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
	redisContainer, err := tcredis.Run(ctx, "redis:6-alpine", network.WithNetwork([]string{"redis"}, nw))
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
	toxiproxyContainer, err := tctoxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0", network.WithNetwork([]string{"toxiproxy"}, nw), testcontainers.WithExposedPorts(proxyPort+"/tcp"))
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
	proxy, err := toxiproxyClient.CreateProxy("redis", "0.0.0.0:"+proxyPort, "redis:6379")
	if err != nil {
		log.Printf("failed to create proxy: %s", err)
		return
	}
	// }
	port, err := toxiproxyContainer.MappedPort(ctx, proxyPort+"/tcp")
	if err != nil {
		log.Printf("failed to get toxiproxy container port: %s", err)
		return
	}
	host, err := toxiproxyContainer.Host(ctx)
	if err != nil {
		log.Printf("failed to get toxiproxy container host: %s", err)
		return
	}
	// createRedisClient {
	options, err := redis.ParseURL(fmt.Sprintf("redis://%s:%s?read_timeout=2s", host, port.Port()))
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
	key, value := fmt.Sprintf("{user.%s}.favoritefood", uuid.NewString()), "Cabbage Biscuits"
	if err := redisCli.Set(ctx, key, value, 2*time.Hour).Err(); err != nil {
		log.Printf("failed to set data: %s", err)
		return
	}
	// addLatencyToxic {
	const latency, jitter = 1_000, 200
	if _, err := proxy.AddToxic("latency_down", "latency", "downstream", 1.0, toxiproxy.Attributes{"latency": latency, "jitter": jitter}); err != nil {
		log.Printf("failed to add toxic: %s", err)
		return
	}
	// }
	start := time.Now()
	savedValue, err := redisCli.Get(ctx, key).Result()
	if err != nil {
		log.Printf("failed to get data: %s", err)
		return
	}
	duration := time.Since(start)
	fmt.Println(savedValue)
	minDuration, maxDuration := (latency-jitter)*time.Millisecond, (latency+jitter)*time.Millisecond
	fmt.Printf("Duration is between %dms and %dms: %v\n", minDuration.Milliseconds(), maxDuration.Milliseconds(), duration >= minDuration && duration <= maxDuration)
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
	redisContainer, err := tcredis.Run(ctx, "redis:6-alpine", network.WithNetwork([]string{"redis"}, nw))
	defer func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	toxiproxyContainer, err := tctoxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0", tctoxiproxy.WithProxy("redis", "redis:6379"), network.WithNetwork([]string{"toxiproxy"}, nw))
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
	proxies, err := toxiproxy.NewClient(toxiURI).Proxies()
	if err != nil {
		log.Printf("failed to get proxies: %s", err)
		return
	}
	proxy := proxies["redis"]
	// readProxiedEndpoint {
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
	key, value := fmt.Sprintf("{user.%s}.favoritefood", uuid.NewString()), "Cabbage Biscuits"
	if err := redisCli.Set(ctx, key, value, 2*time.Hour).Err(); err != nil {
		log.Printf("failed to set data: %s", err)
		return
	}
	if err := proxy.Disable(); err != nil {
		log.Printf("failed to disable proxy: %s", err)
		return
	}
	savedValue, err := redisCli.Get(ctx, key).Result()
	if err == nil {
		log.Printf("proxy is disabled, but we got data")
		return
	}
	fmt.Println(savedValue)
	if err := proxy.Enable(); err != nil {
		log.Printf("failed to enable proxy: %s", err)
		return
	}
	savedValue, err = redisCli.Get(ctx, key).Result()
	if err != nil {
		log.Printf("failed to get data: %s", err)
		return
	}
	fmt.Println(savedValue)
	// Output:
	//
	// Cabbage Biscuits
}
