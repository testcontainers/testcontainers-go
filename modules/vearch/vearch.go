package vearch

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// VearchContainer represents the Vearch container type used in the module
type VearchContainer struct {
	testcontainers.Container
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Vearch container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*VearchContainer, error) {
	return Run(ctx, "vearch/vearch:3.5.1", opts...)
}

// Run creates an instance of the Vearch container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*VearchContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("8817/tcp", "9001/tcp"),
		testcontainers.WithCmd("-conf=/vearch/config.toml", "all"),
		testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.Privileged = true
		}),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      "config.toml",
			ContainerFilePath: "/vearch/config.toml",
			FileMode:          0o666,
		}),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("8817/tcp").WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("9001/tcp").WithStartupTimeout(5*time.Second),
		),
	}

	ctr, err := testcontainers.Run(ctx, img, append(moduleOpts, opts...)...)
	var c *VearchContainer
	if ctr != nil {
		c = &VearchContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run vearch: %w", err)
	}

	return c, nil
}

// RESTEndpoint returns the REST endpoint of the Vearch container
func (c *VearchContainer) RESTEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "8817/tcp", "http")
}
