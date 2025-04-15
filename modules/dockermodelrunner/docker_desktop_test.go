package dockermodelrunner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
)

// skipIfDockerDesktopNotRunning skips the test if Docker Desktop is not running,
// using the testing library and the log.TestLogger of Testcontainers
func skipIfDockerDesktopNotRunning(t *testing.T) {
	t.Helper()
	isDDRunning, err := isDockerDesktopRunning(log.TestLogger(t))
	require.NoError(t, err)

	if !isDDRunning {
		t.Skipf("Skipping because Docker Desktop is not running")
	}
}

// isDockerDesktopRunning checks if Docker Desktop is running.
func isDockerDesktopRunning(l log.Logger) (bool, error) {
	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to create docker client: %w", err)
	}

	info, err := cli.Info(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to get docker info: %w", err)
	}

	if info.OperatingSystem == "Docker Desktop" {
		return true, nil
	}

	l.Printf("Skipping because Docker Desktop is not running")
	return false, nil
}
