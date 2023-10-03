package k6

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// K6Container represents the K6 container type used in the module
type K6Container struct {
	testcontainers.Container
}


// WithTestScript mounts the given script into the ./test directory in the container
// and passes it to k6 as the test to run.
// The path to the script must be an absolute path
func WithTestScript(scriptPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		script := filepath.Base(scriptPath)
		target := fmt.Sprintf("/tests/%s", script)
		mount := testcontainers.ContainerMount{
			Source: testcontainers.GenericBindMountSource{
			    HostPath: scriptPath,
			},
			Target: testcontainers.ContainerMountTarget(target),
		}
		req.Mounts = append(req.Mounts, mount)
		req.Cmd = append(req.Cmd, target)
	}
}

// WithOptions pass the given options to the k6 run command
func WithOptions(options...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Cmd = append(req.Cmd, options...)
	}
}

// RunContainer creates an instance of the K6 container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*K6Container, error) {
	req := testcontainers.ContainerRequest{
		Image: "szkiba/k6x",
		Cmd:   []string{"run"},
		WaitingFor: wait.ForExit(),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &K6Container{Container: container}, nil
}
