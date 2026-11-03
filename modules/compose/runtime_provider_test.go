package compose

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

type composeRuntimeFunc func(context.Context) (testcontainers.RuntimeConfig, error)

func (f composeRuntimeFunc) Initialize(ctx context.Context) (testcontainers.RuntimeConfig, error) {
	return f(ctx)
}

// Runtime selection is process global, so every case runs in a fresh process.
func TestComposeRuntimeRouting(t *testing.T) {
	if scenario := os.Getenv("TEST_COMPOSE_RUNTIME_SCENARIO"); scenario != "" {
		runComposeRuntimeScenario(t, scenario)
		return
	}

	for _, scenario := range []string{"managed", "default", "local"} {
		t.Run(scenario, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestComposeRuntimeRouting$")
			cmd.Env = append(os.Environ(), "TEST_COMPOSE_RUNTIME_SCENARIO="+scenario)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, string(output))
		})
	}
}

func runComposeRuntimeScenario(t *testing.T, scenario string) {
	t.Helper()

	var initialized atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !initialized.Load() {
			http.Error(w, "Docker request preceded runtime initialization", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("API-Version", "1.45")
		switch {
		case strings.HasSuffix(r.URL.Path, "/_ping"):
			_, _ = w.Write([]byte("OK"))
		case strings.HasSuffix(r.URL.Path, "/info"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ID":"compose-test","ServerVersion":"29.0.0","OSType":"linux","OperatingSystem":"Linux"}`))
		default:
			http.Error(w, "unexpected Docker request: "+r.URL.Path, http.StatusNotFound)
		}
	}))
	defer server.Close()

	endpoint := "tcp://" + strings.TrimPrefix(server.URL, "http://")
	t.Setenv("TESTCONTAINERS_PROVIDER", "docker")
	if scenario == "default" {
		t.Setenv("DOCKER_HOST", endpoint)
	} else {
		require.NoError(t, testcontainers.RegisterProvider("sbx", func() testcontainers.RuntimeProvider {
			return composeRuntimeFunc(func(context.Context) (testcontainers.RuntimeConfig, error) {
				initialized.Store(true)
				return testcontainers.RuntimeConfig{
					DockerHost:       endpoint,
					DockerSocketPath: "/var/run/docker.sock",
				}, nil
			})
		}))
		require.NoError(t, core.SelectRuntime("sbx"))
		t.Setenv("DOCKER_HOST", "tcp://127.0.0.1:1")
		t.Setenv("DOCKER_CONTEXT", "missing-context")
		t.Setenv("DOCKER_TLS_VERIFY", "1")
		t.Setenv("DOCKER_CERT_PATH", filepath.Join(t.TempDir(), "missing-certs"))
	}
	if scenario == "default" {
		initialized.Store(true)
	}

	if scenario == "local" {
		local := NewLocalDockerCompose([]string{filepath.Join(t.TempDir(), "compose.yml")}, "local")
		require.True(t, initialized.Load())
		require.ErrorContains(t, local.Invoke().Error, "use NewDockerCompose")
		require.ErrorContains(t, local.Down().Error, "use NewDockerCompose")
		return
	}

	stackFile := filepath.Join(t.TempDir(), "compose.yml")
	require.NoError(t, os.WriteFile(stackFile, []byte("services:\n  hello:\n    image: hello-world\n"), 0o600))
	stack, err := NewDockerCompose(stackFile)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stack.Close()) })
	require.True(t, initialized.Load())
	require.Equal(t, endpoint, stack.dockerCli.Client().DaemonHost())
	require.Equal(t, endpoint, stack.dockerClient.DaemonHost())
	require.Equal(t, endpoint, stack.provider.Client().DaemonHost())
	if scenario == "managed" {
		require.Equal(t, "default", stack.dockerCli.CurrentContext())
		require.Equal(t, endpoint, stack.dockerCli.DockerEndpoint().Host)
	}
}
