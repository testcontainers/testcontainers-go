package ollama

import (
	"context"

	"github.com/docker/docker/api/types/container"

	"github.com/testcontainers/testcontainers-go"
)

var noopCustomizeRequestOption = func(req *testcontainers.GenericContainerRequest) error { return nil }

// withGpu requests a GPU for the container, which could improve performance for some models.
// This option will be automatically added to the Ollama container to check if the host supports nvidia.
func withGpu() testcontainers.CustomizeRequestOption {
	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	if err != nil {
		return noopCustomizeRequestOption
	}

	info, err := cli.Info(context.Background())
	if err != nil {
		return noopCustomizeRequestOption
	}

	// if the Runtime does not support nvidia, we don't need to request a GPU
	if _, ok := info.Runtimes["nvidia"]; !ok {
		return noopCustomizeRequestOption
	}

	return testcontainers.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
		hostConfig.DeviceRequests = []container.DeviceRequest{
			{
				Count:        -1,
				Capabilities: [][]string{{"gpu"}},
			},
		}
	})
}

// WithUseLocal starts a local Ollama process with the given environment in
// KEY=VALUE for instead of a Docker container which can be more performant
// as it has direct access to the GPU.
// By default OLLAMA_HOST is set to localhost:0 to avoid port conflicts.
func WithUseLocal(envKeyValues ...string) *localProcess {
	sessionID := testcontainers.SessionID()
	return &localProcess{
		sessionID: sessionID,
		logName:   localNamePrefix + "-" + sessionID + ".log",
		env:       envKeyValues,
	}
}
