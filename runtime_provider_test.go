package testcontainers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type testRuntimeProvider struct {
	configuration RuntimeConfig
}

func (p testRuntimeProvider) Initialize(context.Context) (RuntimeConfig, error) {
	return p.configuration, nil
}

// Each subprocess owns the process-wide provider and Docker discovery caches.
func TestRuntimeProviderClient(t *testing.T) {
	mode := os.Getenv("TCGO_RUNTIME_TEST")
	if mode == "" {
		binary, err := os.Executable()
		require.NoError(t, err)
		for _, mode := range []string{"healthy", "health-helper", "unavailable", "unregistered"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.CommandContext(t.Context(), binary, "-test.run=^TestRuntimeProviderClient$")
				cmd.Env = append(os.Environ(), "TCGO_RUNTIME_TEST="+mode)
				out, err := cmd.CombinedOutput()
				require.NoError(t, err, "%s", out)
			})
		}
		return
	}

	t.Setenv("TESTCONTAINERS_PROVIDER", "test-runtime")
	if mode == "unregistered" {
		_, err := ProviderDefault.GetProvider()
		require.ErrorContains(t, err, `provider "test-runtime" is not registered`)
		return
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("API-Version", "1.56")
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.WriteHeader(http.StatusOK)
			return
		}
		if mode == "unavailable" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"selected runtime is unavailable"}`))
			return
		}
		_, _ = w.Write([]byte(`{"OperatingSystem":"test runtime","OSType":"linux"}`))
	}))
	t.Cleanup(server.Close)
	want := RuntimeConfig{
		DockerHost:           "tcp://" + strings.TrimPrefix(server.URL, "http://"),
		DockerSocketPath:     "/inner/docker.sock",
		PublishedHost:        "127.0.0.1",
		DefaultBridgeNetwork: "runtime-network",
	}
	require.NoError(t, RegisterProvider("test-runtime", func() RuntimeProvider {
		return testRuntimeProvider{configuration: want}
	}))
	if mode == "health-helper" {
		t.Cleanup(func() {
			require.False(t, t.Skipped(), "the health helper must check the configured provider")
		})
		SkipIfProviderIsNotHealthy(t)
	}
	p, err := NewDockerProvider()
	if mode == "unavailable" {
		require.ErrorContains(t, err, "selected runtime is unavailable")
		require.Nil(t, p)
		return
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = p.Close() })
	want.Name = "test-runtime"
	require.Equal(t, want, p.RuntimeConfig())
	require.Equal(t, "runtime-network", p.defaultBridgeNetworkName)
	host, err := p.DaemonHost(t.Context())
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", host)
	_, err = ProviderDocker.GetProvider()
	require.ErrorContains(t, err, "cannot switch")
}
