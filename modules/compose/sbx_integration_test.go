package compose

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// The integration harness must import modules/sbx to register the provider.
func TestSBXComposeIntegration(t *testing.T) {
	if os.Getenv("TESTCONTAINERS_SBX_INTEGRATION") != "1" {
		t.Skip("set TESTCONTAINERS_SBX_INTEGRATION=1 to run sbx Compose integration")
	}
	require.True(t, core.HasRegisteredProvider("sbx"), "the integration harness must import modules/sbx")
	require.NoError(t, core.SelectRuntime("sbx"))

	composeFile := filepath.Join(t.TempDir(), "compose.yml")
	require.NoError(t, os.WriteFile(composeFile, []byte("services:\n  nginx:\n    image: nginx:stable-alpine\n    ports:\n      - '80'\n"), 0o600))
	stack, err := NewDockerCompose(composeFile)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stack.Close()) })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	require.NoError(t, stack.Up(ctx, Wait(true)))
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Minute)
		defer cleanupCancel()
		require.NoError(t, stack.Down(cleanupCtx))
	})

	container, err := stack.ServiceContainer(ctx, "nginx")
	require.NoError(t, err)
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "80/tcp")
	require.NoError(t, err)
	client := &http.Client{Timeout: 2 * time.Second}
	t.Cleanup(client.CloseIdleConnections)
	require.Eventually(t, func() bool {
		resp, err := client.Get(fmt.Sprintf("http://%s:%s/", host, port.Port()))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 30*time.Second, 500*time.Millisecond)
}
