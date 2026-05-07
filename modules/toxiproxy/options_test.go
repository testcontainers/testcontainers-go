package toxiproxy

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestWithProxy(t *testing.T) {
	t.Run("upstream-is-valid", func(t *testing.T) {
		opt := WithProxy("redis", "redis:6379")

		var opts options
		err := opt(&opts)
		require.NoError(t, err)

		require.Equal(t, "redis", opts.proxies[0].Name)
	})

	t.Run("upstream-is-invalid", func(t *testing.T) {
		opt := WithProxy("redis", "redis:6379:80")

		var opts options
		err := opt(&opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "split hostPort")
	})

	t.Run("upstream-is-invalid-port", func(t *testing.T) {
		opt := WithProxy("redis", "redis:abcde")

		var opts options
		err := opt(&opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "atoi")
	})
}

func TestWithConfigFile(t *testing.T) {
	t.Run("config-file-is-valid", func(t *testing.T) {
		opt := WithConfigFile(strings.NewReader(`[]`))

		var opts options
		err := opt(&opts)
		require.NoError(t, err)
		require.Empty(t, opts.proxies)
	})

	t.Run("config-file-is-invalid", func(t *testing.T) {
		opt := WithConfigFile(strings.NewReader(`invalid`))

		var opts options
		err := opt(&opts)
		require.Error(t, err)
	})

	t.Run("config-file-is-valid-with-proxy", func(t *testing.T) {
		opt := WithConfigFile(strings.NewReader(`[{"name": "redis", "listen": "0.0.0.0:8666", "upstream": "redis:6379", "enabled": true}]`))

		var opts options
		err := opt(&opts)
		require.NoError(t, err)
		require.Equal(t, "redis", opts.proxies[0].Name)
		require.Equal(t, "redis:6379", opts.proxies[0].Upstream)
		require.Empty(t, opts.proxies[0].Listen) // listen is set by the container, as it knows the port
		require.True(t, opts.proxies[0].Enabled)
	})

	t.Run("config-file-is-valid-with-multiple-proxies", func(t *testing.T) {
		opt := WithConfigFile(strings.NewReader(`[{"name": "redis", "listen": "0.0.0.0:8666", "upstream": "redis:6379", "enabled": true}, {"name": "redis2", "listen": "0.0.0.0:8667", "upstream": "redis2:6379", "enabled": true}]`))

		var opts options
		err := opt(&opts)
		require.NoError(t, err)
		require.Len(t, opts.proxies, 2)
		require.Equal(t, "redis", opts.proxies[0].Name)
		require.Equal(t, "redis:6379", opts.proxies[0].Upstream)
		require.Empty(t, opts.proxies[0].Listen) // listen is set by the container, as it knows the port
		require.True(t, opts.proxies[0].Enabled)
		require.Equal(t, "redis2", opts.proxies[1].Name)
		require.Equal(t, "redis2:6379", opts.proxies[1].Upstream)
		require.Empty(t, opts.proxies[1].Listen) // listen is set by the container, as it knows the port
	})

	t.Run("config-file-is-valid-with-invalid-proxy", func(t *testing.T) {
		opt := WithConfigFile(strings.NewReader(`[{"name": "redis", "listen": "0.0.0.0:8666", "upstream": "redis2:6379:80", "enabled": true}]`))

		var opts options
		err := opt(&opts)
		require.Error(t, err)
	})
}

func TestRun_withConfigFile_and_proxy(t *testing.T) {
	configContent := `[
    {
        "name": "redis",
        "listen": "0.0.0.0:8666",
        "upstream": "redis:6379",
        "enabled": true
    }
]`

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

	toxiproxyContainer, err := Run(
		ctx,
		"ghcr.io/shopify/toxiproxy:2.12.0",
		// the config file defines a proxy named "redis"
		WithConfigFile(strings.NewReader(configContent)),
		// this proxy will be added to the existing proxies
		WithProxy("redis2", "redis2:6379"),
		network.WithNetwork([]string{"toxiproxy"}, nw),
	)
	testcontainers.CleanupContainer(t, toxiproxyContainer)
	require.NoError(t, err)

	t.Run("config-file/exists", func(t *testing.T) {
		rc, err := toxiproxyContainer.CopyFileFromContainer(ctx, "/tmp/tc-toxiproxy.json")
		require.NoError(t, err)

		// check that the config file contains two proxies
		var config []proxy
		err = json.NewDecoder(rc).Decode(&config)
		require.NoError(t, err)
		require.Len(t, config, 2)

		require.Contains(t, config, proxy{
			Name:     "redis",
			Listen:   "0.0.0.0:8666",
			Upstream: "redis:6379",
			Enabled:  true,
		})

		require.Contains(t, config, proxy{
			Name:     "redis2",
			Listen:   "0.0.0.0:8667",
			Upstream: "redis2:6379",
			Enabled:  true,
		})
	})

	t.Run("proxied-endpoint/exists", func(t *testing.T) {
		host, port, err := toxiproxyContainer.ProxiedEndpoint(8666)
		require.NoError(t, err)
		require.NotEmpty(t, host)
		require.NotEmpty(t, port)

		host, port, err = toxiproxyContainer.ProxiedEndpoint(8667)
		require.NoError(t, err)
		require.NotEmpty(t, host)
		require.NotEmpty(t, port)
	})

	t.Run("proxied-endpoint/does-not-exist", func(t *testing.T) {
		host, port, err := toxiproxyContainer.ProxiedEndpoint(9999)
		require.Error(t, err)
		require.Empty(t, host)
		require.Empty(t, port)
	})
}
