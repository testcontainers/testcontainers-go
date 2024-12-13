package ollama

import (
	"context"

	"github.com/docker/docker/api/types/container"

	"github.com/testcontainers/testcontainers-go"
)

var noopCustomizeRequestOption = func(req *testcontainers.GenericContainerRequest) error { return nil }

// withGpu requests a GPU for the container, which could improve performance for some models.
// This option will be automaticall added to the Ollama container to check if the host supports nvidia.
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

var _ testcontainers.ContainerCustomizer = (*UseLocal)(nil)

// UseLocal will use the local Ollama instance instead of pulling the Docker image.
type UseLocal struct {
	env map[string]string
}

// WithUseLocal the module will use the local Ollama instance instead of pulling the Docker image.
// Pass the environment variables you need to set for the Ollama binary to be used,
// in the format of "KEY=VALUE". KeyValue pairs with the wrong format will cause an error.
func WithUseLocal(keyVal map[string]string) UseLocal {
	return UseLocal{env: keyVal}
}

// Customize implements the ContainerCustomizer interface, taking the key value pairs
// and setting them as environment variables for the Ollama binary.
// In the case of an invalid key value pair, an error is returned.
func (u UseLocal) Customize(req *testcontainers.GenericContainerRequest) error {
	if len(u.env) == 0 {
		return nil
	}

	return testcontainers.WithEnv(u.env)(req)
}
