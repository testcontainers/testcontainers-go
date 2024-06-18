package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"

	"github.com/testcontainers/testcontainers-go/auth"
	tcimage "github.com/testcontainers/testcontainers-go/image"
	"github.com/testcontainers/testcontainers-go/internal/core"
	tcnetwork "github.com/testcontainers/testcontainers-go/internal/core/network"
	tclog "github.com/testcontainers/testcontainers-go/log"
	tcmount "github.com/testcontainers/testcontainers-go/mount"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Request struct {
	FromDockerfile
	Logger                  tclog.Logging
	HostAccessPorts         []int
	Image                   string
	ImageSubstitutors       []tcimage.Substitutor
	Entrypoint              []string
	Env                     map[string]string
	ExposedPorts            []string // allow specifying protocol info
	Cmd                     []string
	Labels                  map[string]string
	Mounts                  tcmount.ContainerMounts
	Tmpfs                   map[string]string
	WaitingFor              wait.Strategy
	Name                    string // for specifying container name
	Hostname                string
	WorkingDir              string                                     // specify the working directory of the container
	Privileged              bool                                       // For starting privileged container
	Networks                []string                                   // for specifying network names
	NetworkAliases          map[string][]string                        // for specifying network aliases
	Files                   []ContainerFile                            // files which will be copied when container starts
	User                    string                                     // for specifying uid:gid
	AlwaysPullImage         bool                                       // Always pull image
	ImagePlatform           string                                     // ImagePlatform describes the platform which the image runs on.
	ShmSize                 int64                                      // Amount of memory shared with the host (in bytes)
	ConfigModifier          func(*container.Config)                    // Modifier for the config before container creation
	HostConfigModifier      func(*container.HostConfig)                // Modifier for the host config before container creation
	EnpointSettingsModifier func(map[string]*network.EndpointSettings) // Modifier for the network settings before container creation
	LifecycleHooks          []LifecycleHooks                           // define hooks to be executed during container lifecycle
	LogConsumerCfg          *tclog.ConsumerConfig                      // define the configuration for the log producer and its log consumer to follow the logs
	Started                 bool                                       // flag to indicate if the container is started after created
	Reuse                   bool                                       // Experimental. flag to indicate if the container should be reused
}

// BuildOptions returns the image build options when building a Docker image from a Dockerfile.
// It will apply some defaults and finally call the BuildOptionsModifier from the FromDockerfile struct,
// if set.
func (r *Request) BuildOptions() (types.ImageBuildOptions, error) {
	buildOptions := types.ImageBuildOptions{
		Remove:      true,
		ForceRemove: true,
	}

	if r.FromDockerfile.BuildOptionsModifier != nil {
		r.FromDockerfile.BuildOptionsModifier(&buildOptions)
	}

	// apply mandatory values after the modifier
	buildOptions.BuildArgs = r.GetBuildArgs()
	buildOptions.Dockerfile = r.GetDockerfile()

	buildContext, err := r.GetContext()
	if err != nil {
		return buildOptions, err
	}
	buildOptions.Context = buildContext

	// Make sure the auth configs from the Dockerfile are set right after the user-defined build options.
	authsFromDockerfile := getAuthConfigsFromDockerfile(r)

	if buildOptions.AuthConfigs == nil {
		buildOptions.AuthConfigs = map[string]registry.AuthConfig{}
	}

	for registry, authConfig := range authsFromDockerfile {
		buildOptions.AuthConfigs[registry] = authConfig
	}

	// make sure the first tag is the one defined in the Request
	tag := fmt.Sprintf("%s:%s", r.GetRepo(), r.GetTag())

	// apply substitutors to the built image
	for _, is := range r.ImageSubstitutors {
		modifiedTag, err := is.Substitute(tag)
		if err != nil {
			return buildOptions, fmt.Errorf("failed to substitute image %s with %s: %w", tag, is.Description(), err)
		}

		if modifiedTag != tag {
			r.Logger.Printf("✍🏼 Replacing image with %s. From: %s to %s\n", is.Description(), tag, modifiedTag)
			tag = modifiedTag
		}
	}

	if len(buildOptions.Tags) > 0 {
		// prepend the tag
		buildOptions.Tags = append([]string{tag}, buildOptions.Tags...)
	} else {
		buildOptions.Tags = []string{tag}
	}

	if !r.ShouldKeepBuiltImage() {
		buildOptions.Labels = core.DefaultLabels(core.SessionID())
	}

	return buildOptions, nil
}

// getAuthConfigsFromDockerfile returns the auth configs to be able to pull from an authenticated docker registry
func getAuthConfigsFromDockerfile(r *Request) map[string]registry.AuthConfig {
	images, err := core.ExtractImagesFromDockerfile(filepath.Join(r.Context, r.GetDockerfile()), r.GetBuildArgs())
	if err != nil {
		return map[string]registry.AuthConfig{}
	}

	authConfigs := map[string]registry.AuthConfig{}
	for _, img := range images {
		registry, authConfig, err := auth.ForDockerImage(context.Background(), img)
		if err != nil {
			continue
		}

		authConfigs[registry] = authConfig
	}

	return authConfigs
}

// GetBuildArgs returns the env args to be used when creating from Dockerfile
func (r *Request) GetBuildArgs() map[string]*string {
	return r.FromDockerfile.BuildArgs
}

// GetContext retrieve the build context for the request
func (r *Request) GetContext() (io.Reader, error) {
	var includes []string = []string{"."}

	if r.ContextArchive != nil {
		return r.ContextArchive, nil
	}

	// always pass context as absolute path
	abs, err := filepath.Abs(r.Context)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path: %w", err)
	}
	r.Context = abs

	dockerIgnoreExists, excluded, err := tcimage.ParseDockerIgnore(abs)
	if err != nil {
		return nil, err
	}

	if dockerIgnoreExists {
		// only add .dockerignore if it exists
		includes = append(includes, ".dockerignore")
	}

	includes = append(includes, r.GetDockerfile())

	buildContext, err := archive.TarWithOptions(
		r.Context,
		&archive.TarOptions{ExcludePatterns: excluded, IncludeFiles: includes},
	)
	if err != nil {
		return nil, err
	}

	return buildContext, nil
}

// GetDockerfile returns the Dockerfile from the Request, defaults to "Dockerfile"
func (r *Request) GetDockerfile() string {
	f := r.FromDockerfile.Dockerfile
	if f == "" {
		return "Dockerfile"
	}

	return f
}

// GetRepo returns the Repo label for image from the Request, defaults to UUID
func (r *Request) GetRepo() string {
	repo := r.FromDockerfile.Repo
	if repo == "" {
		return uuid.NewString()
	}

	return strings.ToLower(repo)
}

// GetTag returns the Tag label for image from the Request, defaults to UUID
func (r *Request) GetTag() string {
	t := r.FromDockerfile.Tag
	if t == "" {
		return uuid.NewString()
	}

	return strings.ToLower(t)
}

func (r *Request) preCreateContainerHook(ctx context.Context, dockerInput *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig) error {
	// prepare mounts
	hostConfig.Mounts = r.Mounts.Prepare()

	endpointSettings := map[string]*network.EndpointSettings{}

	// #248: Docker allows only one network to be specified during container creation
	// If there is more than one network specified in the request container should be attached to them
	// once it is created. We will take a first network if any specified in the request and use it to create container
	if len(r.Networks) > 0 {
		attachContainerTo := r.Networks[0]

		nw, err := tcnetwork.GetByName(ctx, attachContainerTo)
		if err == nil {
			aliases := []string{}
			if _, ok := r.NetworkAliases[attachContainerTo]; ok {
				aliases = r.NetworkAliases[attachContainerTo]
			}
			endpointSetting := network.EndpointSettings{
				Aliases:   aliases,
				NetworkID: nw.ID,
			}
			endpointSettings[attachContainerTo] = &endpointSetting
		}
	}

	if r.ConfigModifier != nil {
		r.ConfigModifier(dockerInput)
	}

	if r.HostConfigModifier != nil {
		r.HostConfigModifier(hostConfig)
	}

	if r.EnpointSettingsModifier != nil {
		r.EnpointSettingsModifier(endpointSettings)
	}

	networkingConfig.EndpointsConfig = endpointSettings

	exposedPorts := r.ExposedPorts
	// this check must be done after the pre-creation Modifiers are called, so the network mode is already set
	if len(exposedPorts) == 0 && !hostConfig.NetworkMode.IsContainer() {
		cli, err := core.NewClient(ctx)
		if err != nil {
			return err
		}
		defer cli.Close()

		img, _, err := cli.ImageInspectWithRaw(ctx, dockerInput.Image)
		if err != nil && client.IsErrNotFound(err) {
			// pull the image in order to have it available for inspection
			if pullErr := tcimage.Pull(ctx, dockerInput.Image, tclog.StandardLogger(), image.PullOptions{}); pullErr != nil {
				return fmt.Errorf("error pulling image %s: %w", dockerInput.Image, pullErr)
			}

			// try to inspect the image again
			img, _, err = cli.ImageInspectWithRaw(ctx, dockerInput.Image)
		}
		if err != nil {
			return err
		}

		for p := range img.Config.ExposedPorts {
			exposedPorts = append(exposedPorts, string(p))
		}
	}

	exposedPortSet, exposedPortMap, err := nat.ParsePortSpecs(exposedPorts)
	if err != nil {
		return err
	}

	dockerInput.ExposedPorts = exposedPortSet

	// only exposing those ports automatically if the container request exposes zero ports and the container does not run in a container network
	if len(exposedPorts) == 0 && !hostConfig.NetworkMode.IsContainer() {
		hostConfig.PortBindings = exposedPortMap
	} else {
		hostConfig.PortBindings = mergePortBindings(hostConfig.PortBindings, exposedPortMap, r.ExposedPorts)
	}

	return nil
}

func (r *Request) Printf(format string, args ...interface{}) {
	if r.Logger != nil {
		r.Logger.Printf(format, args...)
	}
}

func (r *Request) ShouldBuildImage() bool {
	return r.FromDockerfile.Context != "" || r.FromDockerfile.ContextArchive != nil
}

func (r *Request) ShouldKeepBuiltImage() bool {
	return r.FromDockerfile.KeepImage
}

func (r *Request) ShouldPrintBuildLog() bool {
	return r.FromDockerfile.PrintBuildLog
}

// Validate ensures that the Request does not have invalid parameters configured to it
// ex. make sure you are not specifying both an image as well as a context
func (r *Request) Validate() error {
	validationMethods := []func() error{
		r.validateContextAndImage,
		r.validateContextOrImageIsSpecified,
		r.validateMounts,
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

func (r *Request) validateContextAndImage() error {
	if r.FromDockerfile.Context != "" && r.Image != "" {
		return errors.New("you cannot specify both an Image and Context in a ContainerRequest")
	}

	return nil
}

func (r *Request) validateContextOrImageIsSpecified() error {
	if r.FromDockerfile.Context == "" && r.FromDockerfile.ContextArchive == nil && r.Image == "" {
		return errors.New("you must specify either a build context or an image")
	}

	return nil
}

// validateMounts ensures that the mounts do not have duplicate targets.
// It will check the Mounts and HostConfigModifier.Binds fields.
func (r *Request) validateMounts() error {
	targets := make(map[string]bool, len(r.Mounts))

	for idx := range r.Mounts {
		m := r.Mounts[idx]
		targetPath := m.Target.Target()
		if targets[targetPath] {
			return fmt.Errorf("%w: %s", tcmount.ErrDuplicateMountTarget, targetPath)
		} else {
			targets[targetPath] = true
		}
	}

	if r.HostConfigModifier == nil {
		return nil
	}

	hostConfig := container.HostConfig{}

	r.HostConfigModifier(&hostConfig)

	if hostConfig.Binds != nil && len(hostConfig.Binds) > 0 {
		for _, bind := range hostConfig.Binds {
			parts := strings.Split(bind, ":")
			if len(parts) != 2 {
				return fmt.Errorf("%w: %s", tcmount.ErrInvalidBindMount, bind)
			}
			targetPath := parts[1]
			if targets[targetPath] {
				return fmt.Errorf("%w: %s", tcmount.ErrDuplicateMountTarget, targetPath)
			} else {
				targets[targetPath] = true
			}
		}
	}

	return nil
}

func mergePortBindings(configPortMap, exposedPortMap nat.PortMap, exposedPorts []string) nat.PortMap {
	if exposedPortMap == nil {
		exposedPortMap = make(map[nat.Port][]nat.PortBinding)
	}

	mappedPorts := make(map[string]struct{}, len(exposedPorts))
	for _, p := range exposedPorts {
		p = strings.Split(p, "/")[0]
		mappedPorts[p] = struct{}{}
	}

	for k, v := range configPortMap {
		if _, ok := mappedPorts[k.Port()]; ok {
			exposedPortMap[k] = v
		}
	}
	return exposedPortMap
}
