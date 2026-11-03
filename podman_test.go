package testcontainers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDefaultNetworkAlias(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNotFound, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/networks"):
					_, _ = w.Write([]byte(`[{"Name":"bridge"},{"Name":"reaper_default"}]`))
				case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/networks/podman"):
					w.WriteHeader(status)
					if status == http.StatusOK {
						_, _ = w.Write([]byte(`{"Name":"podman","Id":"native-network"}`))
					} else {
						_, _ = w.Write([]byte(`{"message":"network lookup failed"}`))
					}
				default:
					t.Errorf("unexpected network request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			t.Cleanup(server.Close)
			cli, err := client.New(client.WithHost(server.URL), client.WithAPIVersion("1.41"))
			require.NoError(t, err)
			t.Cleanup(func() { _ = cli.Close() })
			provider := &DockerProvider{
				DockerProviderOptions: &DockerProviderOptions{
					GenericProviderOptions:   &GenericProviderOptions{},
					defaultBridgeNetworkName: Podman,
				},
				client: &DockerClient{Client: cli},
			}
			network, err := provider.ensureDefaultNetwork(t.Context())
			if status == http.StatusInternalServerError {
				require.ErrorContains(t, err, "network lookup failed")
				require.Empty(t, provider.defaultNetwork, "a failed lookup must not cache the fallback")
				return
			}
			require.NoError(t, err)
			if status == http.StatusNotFound {
				require.Equal(t, ReaperDefault, network)
			} else {
				require.Equal(t, Podman, network)
			}
		})
	}
}

func testDefaultBridge(t *testing.T) string {
	t.Helper()
	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close() })
	bridge, err := provider.ensureDefaultNetwork(t.Context())
	require.NoError(t, err)
	return bridge
}

func skipIfRootless(t *testing.T) {
	t.Helper()
	cli, err := NewDockerClientWithOpts(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { _ = cli.Close() })
	info, err := cli.Info(t.Context(), client.InfoOptions{})
	require.NoError(t, err)
	for _, option := range info.Info.SecurityOptions {
		if strings.Contains(option, "rootless") {
			t.Skip("requires a rootful runtime")
		}
	}
}

// The CI harness runs each mode in a fresh process, without the Podman module
// import or a Go workspace. The ordinary core suite uses the registered module.
func TestPodmanCompatibility(t *testing.T) {
	mode := os.Getenv("TCGO_PODMAN_COMPAT")
	if mode == "" {
		t.Skip("run by the native Podman compatibility job")
	}
	require.False(t, core.HasRegisteredProvider("podman"))
	require.Empty(t, os.Getenv("TESTCONTAINERS_PROVIDER"))

	selection := ProviderDefault
	switch mode {
	case "auto":
		require.Contains(t, os.Getenv("DOCKER_HOST"), "podman.sock")
	case "explicit":
		selection = ProviderPodman
	default:
		t.Fatalf("unknown compatibility mode %q", mode)
	}
	provider, err := selection.GetProvider()
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close() })
	p, ok := provider.(*DockerProvider)
	require.True(t, ok)
	require.Equal(t, "podman", p.RuntimeConfig().Name)
	require.Equal(t, Podman, p.defaultBridgeNetworkName)

	ctx := t.Context()
	ctr, err := provider.RunContainer(ctx, ContainerRequest{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForHTTP("/").WithPort(nginxDefaultPort),
	})
	CleanupContainer(t, ctr)
	require.NoError(t, err)
	networks, err := ctr.Networks(ctx)
	require.NoError(t, err)
	bridge, err := p.ensureDefaultNetwork(ctx)
	require.NoError(t, err)
	require.Contains(t, networks, bridge)
	endpoint, err := ctr.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)
	response, err := http.Get(endpoint)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.NoError(t, TerminateContainer(ctr))
	_, err = p.client.ContainerInspect(ctx, ctr.GetContainerID(), client.ContainerInspectOptions{})
	require.True(t, errdefs.IsNotFound(err), "container still exists: %v", err)
}
