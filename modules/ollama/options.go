package ollama

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
)

var noopCustomizeRequestOption = func(req *testcontainers.GenericContainerRequest) {}

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

// WithModel will run the given model, without any prompt.
// If Ollama is not able to run the given model, it will fail to initialise.
func WithModel(model string) testcontainers.CustomizeRequestOption {
	pullCmds := []string{"ollama", "pull", model}
	runCmds := []string{"ollama", "run", model}

	return func(req *testcontainers.GenericContainerRequest) {
		modelLifecycleHook := testcontainers.ContainerLifecycleHooks{
			PostReadies: []testcontainers.ContainerHook{
				func(ctx context.Context, c testcontainers.Container) error {
					_, _, err := c.Exec(ctx, pullCmds, exec.Multiplexed())
					if err != nil {
						return fmt.Errorf("failed to pull model %s: %w", model, err)
					}

					_, r, err := c.Exec(ctx, runCmds, exec.Multiplexed())
					if err != nil {
						return fmt.Errorf("failed to run model %s: %w", model, err)
					}

					bs, err := io.ReadAll(r)
					if err != nil {
						return fmt.Errorf("failed to run %s model: %w", model, err)
					}

					stdOutput := string(bs)
					if strings.Contains(stdOutput, "Error: pull model manifest: file does not exist") {
						return fmt.Errorf("failed to run %s model [%v]: %s", model, runCmds, stdOutput)
					}

					return nil
				},
			},
		}

		req.LifecycleHooks = append(req.LifecycleHooks, modelLifecycleHook)
	}
}
