package dockermcpgateway_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	dmcpg "github.com/testcontainers/testcontainers-go/modules/dockermcpgateway"
)

func TestDockerMCPGateway(t *testing.T) {
	ctx := t.Context()

	ctr, err := dmcpg.Run(ctx, "docker/mcp-gateway:latest")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	require.Empty(t, ctr.Tools())
}

func TestDockerMCPGateway_withServerAndTools(t *testing.T) {
	ctx := t.Context()

	ctr, err := dmcpg.Run(
		ctx, "docker/mcp-gateway:latest",
		dmcpg.WithTools("curl", []string{"curl"}),
		dmcpg.WithTools("duckduckgo", []string{"fetch_content", "search"}),
		dmcpg.WithTools("github-official", []string{"add_issue_comment"}),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	require.Len(t, ctr.Tools(), 3)

	for server, tools := range ctr.Tools() {
		switch server {
		case "curl":
			require.Equal(t, []string{"curl"}, tools)
		case "duckduckgo":
			require.ElementsMatch(t, []string{"fetch_content", "search"}, tools)
		case "github-official":
			require.Equal(t, []string{"add_issue_comment"}, tools)
		default:
			t.Errorf("unexpected server: %s", server)
		}
	}
}

func TestDockerMCPGateway_withSecret(t *testing.T) {
	ctx := t.Context()

	ctr, err := dmcpg.Run(
		ctx, "docker/mcp-gateway:latest",
		dmcpg.WithSecret("github.personal_access_token", "test_token"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	r, err := ctr.CopyFileFromContainer(ctx, "/testcontainers/app/secrets")
	require.NoError(t, err)

	bytes, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, "github.personal_access_token=test_token\n", string(bytes))
}

func TestDockerMCPGateway_withSecrets(t *testing.T) {
	ctx := t.Context()

	ctr, err := dmcpg.Run(
		ctx, "docker/mcp-gateway:latest",
		dmcpg.WithSecrets(map[string]string{
			"github.personal_access_token": "test_token",
			"another.secret":               "another_value",
		}),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	r, err := ctr.CopyFileFromContainer(ctx, "/testcontainers/app/secrets")
	require.NoError(t, err)

	bytes, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, "github.personal_access_token=test_token\nanother.secret=another_value\n", string(bytes))
}
