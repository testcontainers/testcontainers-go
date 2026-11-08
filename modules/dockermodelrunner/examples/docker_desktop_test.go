package dockermodelrunner_test

import (
	"context"
	"fmt"

	"github.com/moby/moby/client"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
)

// isDockerDesktopRunning checks if Docker Desktop is running.
func isDockerDesktopRunning(l log.Logger) (bool, error) {
	apiClient, err := testcontainers.NewDockerClientWithOpts(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to create docker client: %w", err)
	}

	res, err := apiClient.Info(context.Background(), client.InfoOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get docker info: %w", err)
	}

	if res.Info.OperatingSystem == "Docker Desktop" {
		return true, nil
	}

	l.Printf("Skipping because Docker Desktop is not running")
	return false, nil
}
