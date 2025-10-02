package dind

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultDockerDaemonPortNumber = "2375"
	defaultDockerDaemonPort       = defaultDockerDaemonPortNumber + "/tcp"
)

// Container represents the Docker in Docker container type used in the module
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the Docker in Docker container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithCmd(
			"dockerd", "-H", "tcp://0.0.0.0:"+defaultDockerDaemonPortNumber, "--tls=false",
		),
		testcontainers.WithEnv(map[string]string{
			"DOCKER_HOST": "tcp://localhost:" + defaultDockerDaemonPortNumber,
		}),
		testcontainers.WithExposedPorts(defaultDockerDaemonPort),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(defaultDockerDaemonPort)),
		testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.Privileged = true
			hc.CgroupnsMode = "host"
			hc.Tmpfs = map[string]string{
				"/run":     "",
				"/var/run": "",
			}
			hc.Mounts = []mount.Mount{}
		}),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run dind: %w", err)
	}

	return c, nil
}

// Host returns the endpoint to connect to the Docker daemon running inside the DinD container.
func (c *Container) Host(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, defaultDockerDaemonPort, "http")
}

// LoadImage loads an image into the DinD container.
// It creates a temporary file to save the image and then copies it to the container.
// This temporary file is deleted after the function returns.
func (c *Container) LoadImage(ctx context.Context, image string) (err error) {
	var provider testcontainers.GenericProvider
	if provider, err = testcontainers.ProviderDocker.GetProvider(); err != nil {
		return fmt.Errorf("get docker provider: %w", err)
	}

	// save image
	imagesTar, err := os.CreateTemp(os.TempDir(), "image*.tar")
	if err != nil {
		return fmt.Errorf("create temporary images file: %w", err)
	}
	defer func() {
		err = errors.Join(err, os.Remove(imagesTar.Name()))
	}()

	if err = provider.SaveImages(ctx, imagesTar.Name(), image); err != nil {
		return fmt.Errorf("save images: %w", err)
	}

	containerPath := "/image/" + filepath.Base(imagesTar.Name())
	if err = c.CopyFileToContainer(ctx, imagesTar.Name(), containerPath, 0o644); err != nil {
		return fmt.Errorf("copy image to container: %w", err)
	}

	if _, _, err = c.Exec(ctx, []string{"docker", "image", "load", "-i", containerPath}); err != nil {
		return fmt.Errorf("import image: %w", err)
	}

	return nil
}
