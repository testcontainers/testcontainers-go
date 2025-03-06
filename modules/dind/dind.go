package dind

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	// containerPorts {
	defaultDockerDaemonPort = "2375/tcp"
)

// DinDContainer represents the Docker in Docker container type used in the module
type DinDContainer struct {
	testcontainers.Container
}

// Run creates an instance of the K3s container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DinDContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
		ExposedPorts: []string{
			defaultDockerDaemonPort,
		},
		Privileged: true,
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.CgroupnsMode = "host"
			hc.Tmpfs = map[string]string{
				"/run":     "",
				"/var/run": "",
			}
			hc.Mounts = []mount.Mount{}
		},
		Cmd: []string{
			"dockerd", "-H", "tcp://0.0.0.0:2375", "--tls=false",
		},
		Env: map[string]string{
			"DOCKER_HOST": "tcp://localhost:2375",
		},
		WaitingFor: wait.ForListeningPort("2375/tcp"),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *DinDContainer
	if container != nil {
		c = &DinDContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

func (c *DinDContainer) Host(ctx context.Context) (string, error) {
	return c.Container.PortEndpoint(ctx, "2375/tcp", "http")
}

// LoadImage loads an image into the DinD container.
func (c *DinDContainer) LoadImage(ctx context.Context, image string) error {
	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		return fmt.Errorf("getting docker provider %w", err)
	}

	// save image
	imagesTar, err := os.CreateTemp(os.TempDir(), "image*.tar")
	if err != nil {
		return fmt.Errorf("creating temporary images file %w", err)
	}
	defer func() {
		err = errors.Join(err, os.Remove(imagesTar.Name())
	}()

	err = provider.SaveImages(context.Background(), imagesTar.Name(), image)
	if err != nil {
		return fmt.Errorf("saving images %w", err)
	}

	containerPath := "/image/" + filepath.Base(imagesTar.Name())
	err = c.Container.CopyFileToContainer(ctx, imagesTar.Name(), containerPath, 0x644)
	if err != nil {
		return fmt.Errorf("copying image to container %w", err)
	}

	_, _, err = c.Container.Exec(ctx, []string{"docker", "image", "import", containerPath, image})
	if err != nil {
		return fmt.Errorf("importing image %w", err)
	}

	return nil
}
