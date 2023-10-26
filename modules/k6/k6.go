package k6

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/mount"

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
		target := "/home/k6x/" + script
		req.Files = append(
			req.Files,
			testcontainers.ContainerFile{
				HostFilePath:      scriptPath,
				ContainerFilePath: target,
				FileMode:          0o644,
			},
		)

		// add script to the k6 run command
		req.Cmd = append(req.Cmd, target)
	}
}

// WithCmdOptions pass the given options to the k6 run command
func WithCmdOptions(options ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Cmd = append(req.Cmd, options...)
	}
}

// SetEnvVar adds a '--env' command-line flag to the k6 command in the container for setting an environment variable for the test script.
func SetEnvVar(variable string, value string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Cmd = append(req.Cmd, "--env", fmt.Sprintf("%s=%s", variable, value))
	}
}

// WithCache sets a volume as a cache for building the k6 binary
// If a volume name is provided in the TC_K6_BUILD_CACHE, this volume is used and it will
// persist across test sessions.
// If no value is provided, a volume is created and automatically deleted when the test session ends.
func WithCache() testcontainers.CustomizeRequestOption {
	var volOptions *mount.VolumeOptions

	cacheVol := os.Getenv("TC_K6_BUILD_CACHE")
	// if no volume is provided, create one and ensure add labels for garbage collection
	if cacheVol == "" {
		cacheVol = fmt.Sprintf("k6-cache-%s", testcontainers.SessionID())
		volOptions = &mount.VolumeOptions{
			Labels: testcontainers.GenericLabels(),
		}
	}

	return func(req *testcontainers.GenericContainerRequest) {
		mount := testcontainers.ContainerMount{
			Source: testcontainers.DockerVolumeMountSource{
				Name:          cacheVol,
				VolumeOptions: volOptions,
			},
			Target: "/cache",
		}
		req.Mounts = append(req.Mounts, mount)
	}
}

// RunContainer creates an instance of the K6 container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*K6Container, error) {
	req := testcontainers.ContainerRequest{
		Image:      "szkiba/k6x:v0.3.1",
		Cmd:        []string{"run"},
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
