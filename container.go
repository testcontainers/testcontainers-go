package testcontainers

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/go-connections/nat"

	tccontainer "github.com/testcontainers/testcontainers-go/container"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/image"
	"github.com/testcontainers/testcontainers-go/log"
	tcmount "github.com/testcontainers/testcontainers-go/mount"
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
	Inspect(context.Context) (*types.ContainerJSON, error)          // get container info
	MappedPort(context.Context, nat.Port) (nat.Port, error)         // get externally mapped port for a container port
	Ports(context.Context) (nat.PortMap, error)                     // Deprecated: Use c.Inspect(ctx).NetworkSettings.Ports instead
	SessionID() string                                              // get session id
	IsRunning() bool
	Start(context.Context) error                                    // start the container
	Stop(context.Context, *time.Duration) error                     // stop the container
	Terminate(context.Context) error                                // terminate the container
	Logs(context.Context) (io.ReadCloser, error)                    // Get logs of the container
	FollowOutput(LogConsumer)                                       // Deprecated: it will be removed in the next major release
	StartLogProducer(context.Context, ...LogProductionOption) error // Deprecated: Use the ContainerRequest instead
	StopLogProducer() error                                         // Deprecated: it will be removed in the next major release
	Name(context.Context) (string, error)                           // Deprecated: Use c.Inspect(ctx).Name instead
	State(context.Context) (*types.ContainerState, error)           // returns container's running state
	Networks(context.Context) ([]string, error)                     // get container networks
	NetworkAliases(context.Context) (map[string][]string, error)    // get container network aliases for a network
	Exec(ctx context.Context, cmd []string, options ...tcexec.ProcessOption) (int, io.Reader, error)
	ContainerIP(context.Context) (string, error)    // get container ip
	ContainerIPs(context.Context) ([]string, error) // get all container IPs
	CopyToContainer(ctx context.Context, fileContent []byte, containerFilePath string, fileMode int64) error
	CopyDirToContainer(ctx context.Context, hostDirPath string, containerParentPath string, fileMode int64) error
	CopyFileToContainer(ctx context.Context, hostFilePath string, containerFilePath string, fileMode int64) error
	CopyFileFromContainer(ctx context.Context, filePath string) (io.ReadCloser, error)
	GetLogProductionErrorChannel() <-chan error
}

// Deprecated: use image.BuildInfo instead
// ImageBuildInfo defines what is needed to build an image
type ImageBuildInfo = image.BuildInfo

// Deprecated: use tccontainer.FromDockerfile instead
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
	// KeepImage describes whether DockerContainer.Terminate should not delete the
	// container image. Useful for images that are built from a Dockerfile and take a
	// long time to build. Keeping the image also Docker to reuse it.
	KeepImage bool
	// BuildOptionsModifier Modifier for the build options before image build. Use it for
	// advanced configurations while building the image. Please consider that the modifier
	// is called after the default build options are set.
	BuildOptionsModifier func(*types.ImageBuildOptions)
}

// Deprecated: use tccontainer.ContainerFile instead
type ContainerFile = tccontainer.ContainerFile

// ContainerRequest represents the parameters used to get a running container
type ContainerRequest struct {
	FromDockerfile
	HostAccessPorts         []int
	Image                   string
	ImageSubstitutors       []ImageSubstitutor
	Entrypoint              []string
	Env                     map[string]string
	ExposedPorts            []string // allow specifying protocol info
	Cmd                     []string
	Labels                  map[string]string
	Mounts                  tcmount.ContainerMounts
	Tmpfs                   map[string]string
	RegistryCred            string // Deprecated: Testcontainers will detect registry credentials automatically
	WaitingFor              wait.Strategy
	Name                    string // for specifying container name
	Hostname                string
	WorkingDir              string                                     // specify the working directory of the container
	ExtraHosts              []string                                   // Deprecated: Use HostConfigModifier instead
	Privileged              bool                                       // For starting privileged container
	Networks                []string                                   // for specifying network names
	NetworkAliases          map[string][]string                        // for specifying network aliases
	NetworkMode             container.NetworkMode                      // Deprecated: Use HostConfigModifier instead
	Resources               container.Resources                        // Deprecated: Use HostConfigModifier instead
	Files                   []tccontainer.ContainerFile                // files which will be copied when container starts
	User                    string                                     // for specifying uid:gid
	SkipReaper              bool                                       // Deprecated: The reaper is globally controlled by the .testcontainers.properties file or the TESTCONTAINERS_RYUK_DISABLED environment variable
	ReaperImage             string                                     // Deprecated: use WithImageName ContainerOption instead. Alternative reaper image
	ReaperOptions           []ContainerOption                          // Deprecated: the reaper is configured at the properties level, for an entire test session
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
	LogConsumerCfg          *log.ConsumerConfig                        // define the configuration for the log producer and its log consumers to follow the logs
}

// containerOptions functional options for a container
type containerOptions struct {
	ImageName           string
	RegistryCredentials string // Deprecated: Testcontainers will detect registry credentials automatically
}

// Deprecated: it will be removed in the next major release
// functional option for setting the reaper image
type ContainerOption func(*containerOptions)

// Deprecated: it will be removed in the next major release
// WithImageName sets the reaper image name
func WithImageName(imageName string) ContainerOption {
	return func(o *containerOptions) {
		o.ImageName = imageName
	}
}

// Deprecated: Testcontainers will detect registry credentials automatically, and it will be removed in the next major release
// WithRegistryCredentials sets the reaper registry credentials
func WithRegistryCredentials(registryCredentials string) ContainerOption {
	return func(o *containerOptions) {
		o.RegistryCredentials = registryCredentials
	}
}
