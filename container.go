package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/wait"
)

// DeprecatedContainer shows methods that were supported before, but are now deprecated
// Deprecated: Use Container
type DeprecatedContainer interface {
	GetHostEndpoint(ctx context.Context, port string) (string, string, error)
	GetIPAddress(ctx context.Context) (string, error)
	LivenessCheckPorts(ctx context.Context) (nat.PortSet, error)
	Terminate(ctx context.Context) error
}

// Container allows getting info about and controlling a single container instance
type Container interface {
	GetContainerID() string                                         // get the container id from the provider
	Endpoint(context.Context, string) (string, error)               // get proto://ip:port string for the first exposed port
	PortEndpoint(context.Context, nat.Port, string) (string, error) // get proto://ip:port string for the given exposed port
	Host(context.Context) (string, error)                           // get host where the container port is exposed
	MappedPort(context.Context, nat.Port) (nat.Port, error)         // get externally mapped port for a container port
	Ports(context.Context) (nat.PortMap, error)                     // get all exposed ports
	SessionID() string                                              // get session id
	IsRunning() bool
	Start(context.Context) error                 // start the container
	Stop(context.Context, *time.Duration) error  // stop the container
	Terminate(context.Context) error             // terminate the container
	Logs(context.Context) (io.ReadCloser, error) // Get logs of the container
	FollowOutput(LogConsumer)
	StartLogProducer(context.Context) error
	StopLogProducer() error
	Name(context.Context) (string, error)                        // get container name
	State(context.Context) (*types.ContainerState, error)        // returns container's running state
	Networks(context.Context) ([]string, error)                  // get container networks
	NetworkAliases(context.Context) (map[string][]string, error) // get container network aliases for a network
	Exec(ctx context.Context, cmd []string, options ...tcexec.ProcessOption) (int, io.Reader, error)
	ContainerIP(context.Context) (string, error)    // get container ip
	ContainerIPs(context.Context) ([]string, error) // get all container IPs
	CopyToContainer(ctx context.Context, fileContent []byte, containerFilePath string, fileMode int64) error
	CopyDirToContainer(ctx context.Context, hostDirPath string, containerParentPath string, fileMode int64) error
	CopyFileToContainer(ctx context.Context, hostFilePath string, containerFilePath string, fileMode int64) error
	CopyFileFromContainer(ctx context.Context, filePath string) (io.ReadCloser, error)
}

// ImageBuildInfo defines what is needed to build an image
type ImageBuildInfo interface {
	GetContext() (io.Reader, error)                 // the path to the build context
	GetDockerfile() string                          // the relative path to the Dockerfile, including the fileitself
	GetRepo() string                                // get repo label for image
	GetTag() string                                 // get tag label for image
	ShouldPrintBuildLog() bool                      // allow build log to be printed to stdout
	ShouldBuildImage() bool                         // return true if the image needs to be built
	GetBuildArgs() map[string]*string               // return the environment args used to build the from Dockerfile
	GetAuthConfigs() map[string]registry.AuthConfig // return the auth configs to be able to pull from an authenticated docker registry
}

// FromDockerfile represents the parameters needed to build an image from a Dockerfile
// rather than using a pre-built one
type FromDockerfile struct {
	Context        string                         // the path to the context of the docker build
	ContextArchive io.Reader                      // the tar archive file to send to docker that contains the build context
	Dockerfile     string                         // the path from the context to the Dockerfile for the image, defaults to "Dockerfile"
	Repo           string                         // the repo label for image, defaults to UUID
	Tag            string                         // the tag label for image, defaults to UUID
	BuildArgs      map[string]*string             // enable user to pass build args to docker daemon
	PrintBuildLog  bool                           // enable user to print build log
	AuthConfigs    map[string]registry.AuthConfig // Deprecated. Testcontainers will detect registry credentials automatically. Enable auth configs to be able to pull from an authenticated docker registry
}

type ContainerFile struct {
	HostFilePath      string
	ContainerFilePath string
	FileMode          int64
}

// ContainerRequest represents the parameters used to get a running container
type ContainerRequest struct {
	FromDockerfile
	Image                   string
	Entrypoint              []string
	Env                     map[string]string
	ExposedPorts            []string // allow specifying protocol info
	Cmd                     []string
	Labels                  map[string]string
	Mounts                  ContainerMounts
	Tmpfs                   map[string]string
	RegistryCred            string // Deprecated: Testcontainers will detect registry credentials automatically
	WaitingFor              wait.Strategy
	Name                    string // for specifying container name
	Hostname                string
	ExtraHosts              []string                                   // Deprecated: Use HostConfigModifier instead
	Privileged              bool                                       // For starting privileged container
	Networks                []string                                   // for specifying network names
	NetworkAliases          map[string][]string                        // for specifying network aliases
	NetworkMode             container.NetworkMode                      // Deprecated: Use HostConfigModifier instead
	Resources               container.Resources                        // Deprecated: Use HostConfigModifier instead
	Files                   []ContainerFile                            // files which will be copied when container starts
	User                    string                                     // for specifying uid:gid
	SkipReaper              bool                                       // Deprecated: The reaper is globally controlled by the .testcontainers.properties file or the TESTCONTAINERS_RYUK_DISABLED environment variable
	ReaperImage             string                                     // Deprecated: use WithImageName ContainerOption instead. Alternative reaper image
	ReaperOptions           []ContainerOption                          // options for the reaper
	AutoRemove              bool                                       // Deprecated: Use HostConfigModifier instead. If set to true, the container will be removed from the host when stopped
	AlwaysPullImage         bool                                       // Always pull image
	ImagePlatform           string                                     // ImagePlatform describes the platform which the image runs on.
	Binds                   []string                                   // Deprecated: Use HostConfigModifier instead
	ShmSize                 int64                                      // Amount of memory shared with the host (in bytes)
	CapAdd                  []string                                   // Deprecated: Use HostConfigModifier instead. Add Linux capabilities
	CapDrop                 []string                                   // Deprecated: Use HostConfigModifier instead. Drop Linux capabilities
	ConfigModifier          func(*container.Config)                    // Modifier for the config before container creation
	HostConfigModifier      func(*container.HostConfig)                // Modifier for the host config before container creation
	EnpointSettingsModifier func(map[string]*network.EndpointSettings) // Modifier for the network settings before container creation
	LifecycleHooks          []ContainerLifecycleHooks                  // define hooks to be executed during container lifecycle
}

// containerOptions functional options for a container
type containerOptions struct {
	ImageName           string
	RegistryCredentials string // Deprecated: Testcontainers will detect registry credentials automatically
}

// functional option for setting the reaper image
type ContainerOption func(*containerOptions)

// WithImageName sets the reaper image name
func WithImageName(imageName string) ContainerOption {
	return func(o *containerOptions) {
		o.ImageName = imageName
	}
}

// Deprecated: Testcontainers will detect registry credentials automatically
// WithRegistryCredentials sets the reaper registry credentials
func WithRegistryCredentials(registryCredentials string) ContainerOption {
	return func(o *containerOptions) {
		o.RegistryCredentials = registryCredentials
	}
}

// Validate ensures that the ContainerRequest does not have invalid parameters configured to it
// ex. make sure you are not specifying both an image as well as a context
func (c *ContainerRequest) Validate() error {
	validationMethods := []func() error{
		c.validateContextAndImage,
		c.validateContextOrImageIsSpecified,
		c.validateMounts,
	}

	var err error
	for _, validationMethod := range validationMethods {
		err = validationMethod()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetContext retrieve the build context for the request
func (c *ContainerRequest) GetContext() (io.Reader, error) {
	if c.ContextArchive != nil {
		return c.ContextArchive, nil
	}

	// always pass context as absolute path
	abs, err := filepath.Abs(c.Context)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path: %w", err)
	}
	c.Context = abs

	buildContext, err := archive.TarWithOptions(c.Context, &archive.TarOptions{})
	if err != nil {
		return nil, err
	}

	return buildContext, nil
}

// GetBuildArgs returns the env args to be used when creating from Dockerfile
func (c *ContainerRequest) GetBuildArgs() map[string]*string {
	return c.FromDockerfile.BuildArgs
}

// GetDockerfile returns the Dockerfile from the ContainerRequest, defaults to "Dockerfile"
func (c *ContainerRequest) GetDockerfile() string {
	f := c.FromDockerfile.Dockerfile
	if f == "" {
		return "Dockerfile"
	}

	return f
}

// GetRepo returns the Repo label for image from the ContainerRequest, defaults to UUID
func (c *ContainerRequest) GetRepo() string {
	r := c.FromDockerfile.Repo
	if r == "" {
		return uuid.NewString()
	}

	return strings.ToLower(r)
}

// GetTag returns the Tag label for image from the ContainerRequest, defaults to UUID
func (c *ContainerRequest) GetTag() string {
	t := c.FromDockerfile.Tag
	if t == "" {
		return uuid.NewString()
	}

	return strings.ToLower(t)
}

// GetAuthConfigs returns the auth configs to be able to pull from an authenticated docker registry
func (c *ContainerRequest) GetAuthConfigs() map[string]registry.AuthConfig {
	images, err := testcontainersdocker.ExtractImagesFromDockerfile(filepath.Join(c.Context, c.GetDockerfile()), c.GetBuildArgs())
	if err != nil {
		return map[string]registry.AuthConfig{}
	}

	authConfigs := map[string]registry.AuthConfig{}
	for _, image := range images {
		registry, authConfig, err := DockerImageAuth(context.Background(), image)
		if err != nil {
			continue
		}

		authConfigs[registry] = authConfig
	}

	return authConfigs
}

func (c *ContainerRequest) ShouldBuildImage() bool {
	return c.FromDockerfile.Context != "" || c.FromDockerfile.ContextArchive != nil
}

func (c *ContainerRequest) ShouldPrintBuildLog() bool {
	return c.FromDockerfile.PrintBuildLog
}

func (c *ContainerRequest) validateContextAndImage() error {
	if c.FromDockerfile.Context != "" && c.Image != "" {
		return errors.New("you cannot specify both an Image and Context in a ContainerRequest")
	}

	return nil
}

func (c *ContainerRequest) validateContextOrImageIsSpecified() error {
	if c.FromDockerfile.Context == "" && c.FromDockerfile.ContextArchive == nil && c.Image == "" {
		return errors.New("you must specify either a build context or an image")
	}

	return nil
}

func (c *ContainerRequest) validateMounts() error {
	targets := make(map[string]bool, len(c.Mounts))

	for idx := range c.Mounts {
		m := c.Mounts[idx]
		targetPath := m.Target.Target()
		if targets[targetPath] {
			return fmt.Errorf("%w: %s", ErrDuplicateMountTarget, targetPath)
		} else {
			targets[targetPath] = true
		}
	}
	return nil
}
