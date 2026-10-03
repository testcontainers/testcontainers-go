package forgejo_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/forgejo"
)

func TestForgejo(t *testing.T) {
	ctx := context.Background()

	ctr, err := forgejo.Run(ctx, "codeberg.org/forgejo/forgejo:11")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// verify connection string returns a valid HTTP URL
	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, connStr)

	// verify the health endpoint is reachable via the connection string
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, connStr+"/api/healthz", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// verify default admin credentials are set
	require.Equal(t, "forgejo-admin", ctr.AdminUsername())
	require.Equal(t, "forgejo-admin", ctr.AdminPassword())
}

func TestForgejoWithAdminCredentials(t *testing.T) {
	ctx := context.Background()

	ctr, err := forgejo.Run(ctx,
		"codeberg.org/forgejo/forgejo:11",
		forgejo.WithAdminCredentials("testuser", "testpassword", "test@example.com"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// verify custom admin credentials are set on the container struct
	require.Equal(t, "testuser", ctr.AdminUsername())
	require.Equal(t, "testpassword", ctr.AdminPassword())

	// verify the API is reachable and admin user can authenticate
	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, connStr+"/api/v1/user", nil)
	require.NoError(t, err)
	req.SetBasicAuth("testuser", "testpassword")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestForgejoSSHEndpoint(t *testing.T) {
	ctx := context.Background()

	ctr, err := forgejo.Run(ctx, "codeberg.org/forgejo/forgejo:11")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	sshStr, err := ctr.SSHConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, sshStr)

	// verify the SSH connection string contains a host and port
	require.Contains(t, sshStr, ":", "SSH connection string should contain host:port")
}
