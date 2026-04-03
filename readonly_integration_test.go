package testcontainers_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestWithReadOnlyRootFilesystem_Integration(t *testing.T) {
	ctx := context.Background()

	// Test that a container with read-only root filesystem cannot write to the root filesystem
	container, err := testcontainers.Run(ctx, "alpine:latest",
		testcontainers.WithReadOnlyRootFilesystem(),
		testcontainers.WithCmd("sh", "-c", "echo 'test' > /test.txt && echo 'success' || echo 'failed'"),
		testcontainers.WithWaitStrategy(wait.ForExit()),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, testcontainers.TerminateContainer(container))
	}()

	// Get the logs to verify the write operation failed
	logs, err := container.Logs(ctx)
	require.NoError(t, err)
	defer logs.Close()

	logBytes, err := io.ReadAll(logs)
	require.NoError(t, err)
	logContent := string(logBytes)

	// The write operation should fail because the root filesystem is read-only
	require.Contains(t, logContent, "failed")
	require.NotContains(t, logContent, "success")

	// Verify the container was actually configured with read-only root filesystem
	inspect, err := container.Inspect(ctx)
	require.NoError(t, err)
	require.True(t, inspect.HostConfig.ReadonlyRootfs)
}

func TestWithReadOnlyRootFilesystem_WithTmpfs_Integration(t *testing.T) {
	ctx := context.Background()

	// Test that a container with read-only root filesystem can still write to tmpfs mounts
	container, err := testcontainers.Run(ctx, "alpine:latest",
		testcontainers.WithReadOnlyRootFilesystem(),
		testcontainers.WithTmpfs(map[string]string{"/tmp": "rw,noexec,nosuid,size=100m"}),
		testcontainers.WithCmd("sh", "-c", "echo 'test' > /tmp/test.txt && echo 'success' || echo 'failed'"),
		testcontainers.WithWaitStrategy(wait.ForExit()),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, testcontainers.TerminateContainer(container))
	}()

	// Get the logs to verify the write operation succeeded in tmpfs
	logs, err := container.Logs(ctx)
	require.NoError(t, err)
	defer logs.Close()

	logBytes, err := io.ReadAll(logs)
	require.NoError(t, err)
	logContent := string(logBytes)

	// The write operation should succeed because /tmp is mounted as tmpfs
	require.Contains(t, logContent, "success")
	require.NotContains(t, logContent, "failed")

	// Verify the container was configured with read-only root filesystem
	inspect, err := container.Inspect(ctx)
	require.NoError(t, err)
	require.True(t, inspect.HostConfig.ReadonlyRootfs)
}