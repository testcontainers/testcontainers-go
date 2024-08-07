package core

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"

	"github.com/testcontainers/testcontainers-go/internal"
	"github.com/testcontainers/testcontainers-go/internal/config"
	tclog "github.com/testcontainers/testcontainers-go/log"
)

const packagePath string = "github.com/testcontainers/testcontainers-go"

var (
	// dockerInfo stores the docker info to be reused in the Info method
	dockerInfo    system.Info
	dockerInfoSet bool
)

// DockerClient is a wrapper around the docker client that is used by testcontainers-go.
// It implements the SystemAPIClient interface in order to cache the docker info and reuse it.
type DockerClient struct {
	client.APIClient // client is embedded into our own client
	logger           tclog.Logging
}

// implements SystemAPIClient interface
var _ client.SystemAPIClient = &DockerClient{}

// Events returns a channel to listen to events that happen to the docker daemon.
func (c *DockerClient) Events(ctx context.Context, options events.ListOptions) (<-chan events.Message, <-chan error) {
	return c.APIClient.Events(ctx, options)
}

// Info returns information about the docker server. The result of Info is cached
// and reused every time Info is called.
// It will also print out the docker server info, and the resolved Docker paths, to the default logger.
func (c *DockerClient) Info(ctx context.Context) (system.Info, error) {
	if dockerInfoSet {
		return dockerInfo, nil
	}

	info, err := c.APIClient.Info(ctx)
	if err != nil {
		return info, fmt.Errorf("failed to retrieve docker info: %w", err)
	}
	dockerInfo = info
	dockerInfoSet = true

	infoMessage := `%v - Connected to docker: 
  Server Version: %v
  API Version: %v
  Operating System: %v
  Total Memory: %v MB%s
  Testcontainers for Go Version: v%s
  Resolved Docker Host: %s
  Resolved Docker Socket Path: %s
  Test SessionID: %s
  Test ProcessID: %s
`

	infoLabels := ""
	if len(dockerInfo.Labels) > 0 {
		infoLabels = `
Labels:`
		for _, lb := range dockerInfo.Labels {
			infoLabels += "\n    " + lb
		}
	}

	c.logger.Printf(infoMessage, packagePath,
		dockerInfo.ServerVersion,
		c.ClientVersion(),
		dockerInfo.OperatingSystem, dockerInfo.MemTotal/1024/1024,
		infoLabels,
		internal.Version,
		ExtractDockerHost(ctx),
		ExtractDockerSocket(ctx),
		SessionID(),
		ProcessID(),
	)

	return dockerInfo, nil
}

// RegistryLogin logs into a Docker registry.
func (c *DockerClient) RegistryLogin(ctx context.Context, auth registry.AuthConfig) (registry.AuthenticateOKBody, error) {
	return c.APIClient.RegistryLogin(ctx, auth)
}

// DiskUsage returns the disk usage of all images.
func (c *DockerClient) DiskUsage(ctx context.Context, options types.DiskUsageOptions) (types.DiskUsage, error) {
	return c.APIClient.DiskUsage(ctx, options)
}

// Ping pings the docker server.
func (c *DockerClient) Ping(ctx context.Context) (types.Ping, error) {
	return c.APIClient.Ping(ctx)
}

var permanentClientErrors = []func(error) bool{
	errdefs.IsNotFound,
	errdefs.IsInvalidParameter,
	errdefs.IsUnauthorized,
	errdefs.IsForbidden,
	errdefs.IsNotImplemented,
	errdefs.IsSystem,
}

func IsPermanentClientError(err error) bool {
	for _, isErrFn := range permanentClientErrors {
		if isErrFn(err) {
			return true
		}
	}
	return false
}

type ContextKey string

const (
	// ClientContextKey is the key used to store the docker client in the context
	ClientContextKey ContextKey = "testcontainers-docker-client"
)

// NewClient returns a new docker client extracting the docker host from the different alternatives
func NewClient(ctx context.Context, ops ...client.Opt) (*DockerClient, error) {
	// check if the context has a client already
	if ctxCli, ok := ctx.Value(ClientContextKey).(*DockerClient); ok {
		return ctxCli, nil
	}

	tcConfig := config.Read()

	dockerHost := ExtractDockerHost(ctx)

	opts := []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}
	if dockerHost != "" {
		opts = append(opts, client.WithHost(dockerHost))

		// for further information, read https://docs.docker.com/engine/security/protect-access/
		if tcConfig.TLSVerify == 1 {
			cacertPath := filepath.Join(tcConfig.CertPath, "ca.pem")
			certPath := filepath.Join(tcConfig.CertPath, "cert.pem")
			keyPath := filepath.Join(tcConfig.CertPath, "key.pem")

			opts = append(opts, client.WithTLSClientConfig(cacertPath, certPath, keyPath))
		}
	}

	opts = append(opts, client.WithHTTPHeaders(
		map[string]string{
			"x-tc-pp":    ProjectPath(),
			"x-tc-sid":   SessionID(),
			"User-Agent": "tc-go/" + internal.Version,
		}),
	)

	// passed options have priority over the default ones
	opts = append(opts, ops...)

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	// check if the engine is available using Info with our wrapper.
	tcClient := &DockerClient{
		APIClient: cli,
		logger:    tclog.StandardLogger(),
	}

	if _, err = tcClient.Info(ctx); err != nil {
		dockerClient, err := client.NewClientWithOpts(opts...)
		if err != nil {
			_ = tcClient.Close()
			return nil, err
		}

		tcClient.APIClient = dockerClient
	}

	return tcClient, nil
}

func NewMockDockerClient(cli client.APIClient) *DockerClient {
	return &DockerClient{
		APIClient: cli,
		logger:    tclog.StandardLogger(),
	}
}
