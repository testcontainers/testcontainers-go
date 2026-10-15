package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
)

type runtimeProviderFunc func(context.Context) (RuntimeConfig, error)

func (f runtimeProviderFunc) Initialize(ctx context.Context) (RuntimeConfig, error) {
	return f(ctx)
}

func TestRuntimeRegistryConcurrentInitialization(t *testing.T) {
	r := newRuntimeRegistry()
	var calls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	require.NoError(t, r.register("sandbox", func() RuntimeProvider {
		return runtimeProviderFunc(func(context.Context) (RuntimeConfig, error) {
			calls.Add(1)
			close(started)
			<-release
			return RuntimeConfig{DockerHost: "tcp://127.0.0.1:12345"}, nil
		})
	}))
	const clients = 32
	results := make(chan error, clients)
	for range clients {
		go func() {
			cfg, err := r.resolve(t.Context(), "sandbox")
			if err == nil && cfg.DockerHost != "tcp://127.0.0.1:12345" {
				err = fmt.Errorf("unexpected endpoint %q", cfg.DockerHost)
			}
			results <- err
		}()
	}
	<-started
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := r.resolve(ctx, "sandbox")
	require.ErrorIs(t, err, context.Canceled)
	close(release)
	for range clients {
		require.NoError(t, <-results)
	}
	require.EqualValues(t, 1, calls.Load())
	require.ErrorContains(t, r.selectProvider("docker"), "already initialized")
	require.ErrorContains(t, r.register("late", func() RuntimeProvider { return nil }), "before Docker discovery")
}

func TestRuntimeRegistryInitializationFailure(t *testing.T) {
	r := newRuntimeRegistry()
	want := errors.New("sandbox startup failed")
	calls := 0
	require.NoError(t, r.register("sandbox", func() RuntimeProvider {
		return runtimeProviderFunc(func(context.Context) (RuntimeConfig, error) {
			calls++
			return RuntimeConfig{}, want
		})
	}))
	for range 2 {
		_, err := r.resolve(t.Context(), "sandbox")
		require.ErrorIs(t, err, want)
	}
	require.Equal(t, 1, calls)
}

func TestRuntimeRegistryCanceledInitialCaller(t *testing.T) {
	r := newRuntimeRegistry()
	var calls atomic.Int32
	started := make(chan context.Context, 1)
	release := make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	defer unblock()
	const endpoint = "tcp://127.0.0.1:12345"
	require.NoError(t, r.register("sandbox", func() RuntimeProvider {
		return runtimeProviderFunc(func(ctx context.Context) (RuntimeConfig, error) {
			calls.Add(1)
			started <- ctx
			<-release
			return RuntimeConfig{DockerHost: endpoint}, ctx.Err()
		})
	}))
	type contextKey struct{}
	ctx, cancel := context.WithTimeout(context.WithValue(t.Context(), contextKey{}, "caller value"), time.Hour)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, err := r.resolve(ctx, "sandbox")
		result <- err
	}()
	startupCtx := <-started
	cancel()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("canceled caller did not stop waiting for runtime startup")
	}
	require.NoError(t, startupCtx.Err())
	_, hasDeadline := startupCtx.Deadline()
	require.False(t, hasDeadline)
	require.Equal(t, "caller value", startupCtx.Value(contextKey{}))
	unblock()
	cfg, err := r.resolve(t.Context(), "sandbox")
	require.NoError(t, err)
	require.Equal(t, endpoint, cfg.DockerHost)
	require.EqualValues(t, 1, calls.Load())
}

func TestRuntimeRegistrySelection(t *testing.T) {
	t.Run("legacy compatibility", func(t *testing.T) {
		for _, name := range []string{"docker", "podman"} {
			r := newRuntimeRegistry()
			cfg, err := r.resolve(t.Context(), name)
			require.NoError(t, err)
			require.Empty(t, cfg.DockerHost)
			if name == "podman" {
				require.Equal(t, "podman", cfg.DefaultBridgeNetwork)
			} else {
				require.Equal(t, "bridge", cfg.DefaultBridgeNetwork)
			}
		}
	})
	t.Run("unknown provider fails", func(t *testing.T) {
		_, err := newRuntimeRegistry().resolve(t.Context(), "missing")
		require.ErrorContains(t, err, `provider "missing" is not registered`)
	})
	t.Run("registration does not initialize", func(t *testing.T) {
		r := newRuntimeRegistry()
		calls := 0
		factory := func() RuntimeProvider {
			calls++
			return runtimeProviderFunc(func(context.Context) (RuntimeConfig, error) {
				return RuntimeConfig{DefaultBridgeNetwork: "module-podman"}, nil
			})
		}
		require.NoError(t, r.register("PoDmAn", factory))
		require.Zero(t, calls)
		require.ErrorContains(t, r.register("podman", factory), "already registered")
		cfg, err := r.resolve(t.Context(), "podman")
		require.NoError(t, err)
		require.Equal(t, "module-podman", cfg.DefaultBridgeNetwork)
		require.Equal(t, 1, calls)
	})
	t.Run("explicit selection precedes configuration", func(t *testing.T) {
		r := newRuntimeRegistry()
		require.NoError(t, r.selectProvider("docker"))
		cfg, err := r.resolve(t.Context(), "missing")
		require.NoError(t, err)
		require.Equal(t, "docker", cfg.Name)
	})
}

func isolateRuntimeRegistry(t *testing.T) {
	t.Helper()
	previous := runtimeProviders
	runtimeProviders = newRuntimeRegistry()
	reset := func() {
		dockerHostOnce = sync.Once{}
		dockerHostCache, dockerHostErrCache = "", nil
		dockerSocketPathOnce = sync.Once{}
		dockerSocketPathCache = ""
		config.Reset()
	}
	reset()
	t.Cleanup(func() {
		runtimeProviders = previous
		reset()
	})
}

func TestRuntimeBeforeEndpointDiscovery(t *testing.T) {
	for _, first := range []string{"host", "socket", "client"} {
		t.Run(first, func(t *testing.T) {
			isolateRuntimeRegistry(t)
			t.Setenv("TESTCONTAINERS_PROVIDER", "sandbox")
			t.Setenv("DOCKER_HOST", "tcp://host-docker.invalid:2376")
			t.Setenv("DOCKER_TLS_VERIFY", "1")
			t.Setenv("DOCKER_CERT_PATH", "/missing/certs")
			t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/host/docker.sock")
			var calls int
			const endpoint = "tcp://127.0.0.1:12345"
			require.NoError(t, RegisterProvider("sandbox", func() RuntimeProvider {
				return runtimeProviderFunc(func(context.Context) (RuntimeConfig, error) {
					calls++
					return RuntimeConfig{DockerHost: endpoint, DockerSocketPath: "/inner/docker.sock"}, nil
				})
			}))
			checkHost := func() {
				host, err := ExtractDockerHost(t.Context())
				require.NoError(t, err)
				require.Equal(t, endpoint, host)
			}
			checkSocket := func() {
				require.Equal(t, "/inner/docker.sock", MustExtractDockerSocket(t.Context()))
			}
			checkClient := func() {
				cli, err := NewClient(t.Context())
				require.NoError(t, err)
				t.Cleanup(func() { _ = cli.Close() })
				require.Equal(t, endpoint, cli.DaemonHost())
			}
			switch first {
			case "host":
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				_, err := ExtractDockerHost(ctx)
				require.ErrorIs(t, err, context.Canceled)
				checkHost()
			case "socket":
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				require.PanicsWithError(t, context.Canceled.Error(), func() {
					MustExtractDockerSocket(ctx)
				})
				checkSocket()
			case "client":
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				_, err := NewClient(ctx)
				require.ErrorIs(t, err, context.Canceled)
				checkClient()
			}
			checkHost()
			checkSocket()
			checkClient()
			require.Equal(t, 1, calls)
		})
	}
}
