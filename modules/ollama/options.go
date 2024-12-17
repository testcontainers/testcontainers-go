package ollama

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

var _ testcontainers.ContainerCustomizer = (*useLocal)(nil)

// useLocal will use the local Ollama instance instead of pulling the Docker image.
type useLocal struct {
	env []string
}

// WithUseLocal the module will use the local Ollama instance instead of pulling the Docker image.
// Pass the environment variables you need to set for the Ollama binary to be used,
// in the format of "KEY=VALUE". KeyValue pairs with the wrong format will cause an error.
func WithUseLocal(values ...string) useLocal {
	return useLocal{env: values}
}

// Customize implements the ContainerCustomizer interface, taking the key value pairs
// and setting them as environment variables for the Ollama binary.
// In the case of an invalid key value pair, an error is returned.
func (u useLocal) Customize(req *testcontainers.GenericContainerRequest) error {
	// Replace the default host port strategy with one that waits for a log entry.
	if err := wait.Walk(&req.WaitingFor, func(w wait.Strategy) error {
		if _, ok := w.(*wait.HostPortStrategy); ok {
			return wait.VisitRemove
		}

		return nil
	}); err != nil {
		return fmt.Errorf("walk strategies: %w", err)
	}

	logStrategy := wait.ForLog(localLogRegex).AsRegexp()
	if req.WaitingFor == nil {
		req.WaitingFor = logStrategy
	} else {
		req.WaitingFor = wait.ForAll(req.WaitingFor, logStrategy)
	}

	osEnv := os.Environ()
	env := make(map[string]string, len(osEnv)+len(u.env)+1)
	// Use a random port to avoid conflicts by default.
	env[localHostVar] = "localhost:0"
	for _, kv := range append(osEnv, u.env...) {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid environment variable: %q", kv)
		}

		env[parts[0]] = parts[1]
	}

	return testcontainers.WithEnv(env)(req)
}
