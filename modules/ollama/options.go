package ollama

import (
	"context"

	"github.com/docker/docker/api/types/container"

	"github.com/testcontainers/testcontainers-go"
)

var noopCustomizeRequestOption = func(_ *testcontainers.GenericContainerRequest) error { return nil }

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
// format KEY=VALUE instead of a Docker container, which can be more performant
// as it has direct access to the GPU.
// By default `OLLAMA_HOST=localhost:0` is set to avoid port conflicts.
//
// When using this option, the container request will be validated to ensure
// that only the options that are compatible with the local process are used.
//
// Supported fields are:
// - [testcontainers.GenericContainerRequest.Started] must be set to true
// - [testcontainers.GenericContainerRequest.ExposedPorts] must be set to ["11434/tcp"]
// - [testcontainers.ContainerRequest.WaitingFor] should not be changed from the default
// - [testcontainers.ContainerRequest.Image] used to determine the local process binary [<path-ignored>/]<binary>[:latest] if not blank.
// - [testcontainers.ContainerRequest.Env] applied to all local process executions
// - [testcontainers.GenericContainerRequest.Logger] is unused
//
// Any other leaf field not set to the type's zero value will result in an error.
func WithUseLocal(envKeyValues ...string) *localProcess {
	sessionID := testcontainers.SessionID()
	return &localProcess{
		sessionID: sessionID,
		logName:   localNamePrefix + "-" + sessionID + ".log",
		env:       envKeyValues,
		binary:    localBinary,
	}
}
