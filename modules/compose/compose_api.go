package compose

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/sync/errgroup"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

type stackUpOptionFunc func(s *stackUpOptions)

func (f stackUpOptionFunc) applyToStackUp(o *stackUpOptions) {
	f(o)
}

// RunServices is comparable to 'docker compose run' as it only creates a subset of containers
// instead of all services defined by the project
func RunServices(serviceNames ...string) StackUpOption {
	return stackUpOptionFunc(func(o *stackUpOptions) {
		o.Services = serviceNames
	})
}

// Deprecated: will be removed in the next major release
// IgnoreOrphans - Ignore legacy containers for services that are not defined in the project
type IgnoreOrphans bool

// Deprecated: will be removed in the next major release
//
//nolint:unused
func (io IgnoreOrphans) applyToStackUp(co *api.CreateOptions, _ *api.StartOptions) {
	co.IgnoreOrphans = bool(io)
}

// Recreate will recreate the containers that are already running
type Recreate string

func (r Recreate) applyToStackUp(o *stackUpOptions) {
	o.Recreate = validateRecreate(string(r))
}

// RecreateDependencies will recreate the dependencies of the services that are already running
type RecreateDependencies string

func (r RecreateDependencies) applyToStackUp(o *stackUpOptions) {
	o.RecreateDependencies = validateRecreate(string(r))
}

func validateRecreate(r string) string {
	switch r {
	case api.RecreateDiverged, api.RecreateForce, api.RecreateNever:
		return r
	default:
		return api.RecreateForce
	}
}

// RemoveOrphans will clean up containers that are not declared on the compose model but own the same labels
type RemoveOrphans bool

func (ro RemoveOrphans) applyToStackUp(o *stackUpOptions) {
	o.RemoveOrphans = bool(ro)
}

func (ro RemoveOrphans) applyToStackDown(o *stackDownOptions) {
	o.RemoveOrphans = bool(ro)
}

// Wait won't return until containers reached the running|healthy state
type Wait bool

func (w Wait) applyToStackUp(o *stackUpOptions) {
	o.Wait = bool(w)
}

type RemoveVolumes bool

func (ro RemoveVolumes) applyToStackDown(o *stackDownOptions) {
	o.Volumes = bool(ro)
}

// RemoveImages used by services
type RemoveImages uint8

func (ri RemoveImages) applyToStackDown(o *stackDownOptions) {
	switch ri {
	case RemoveImagesAll:
		o.Images = "all"
	case RemoveImagesLocal:
		o.Images = "local"
	}
}

type ComposeStackReaders []io.Reader

func (r ComposeStackReaders) applyToComposeStack(o *composeStackOptions) error {
	f := make([]string, len(r))
	baseName := "docker-compose-%d.yml"
	for i, reader := range r {
		tmp := os.TempDir()
		tmp = filepath.Join(tmp, strconv.FormatInt(time.Now().UnixNano(), 10))
		err := os.MkdirAll(tmp, 0o755)
		if err != nil {
			return fmt.Errorf("create temporary directory: %w", err)
		}

		name := fmt.Sprintf(baseName, i)

		bs, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("read from reader: %w", err)
		}

		err = os.WriteFile(filepath.Join(tmp, name), bs, 0o644)
		if err != nil {
			return fmt.Errorf("write to temporary file: %w", err)
		}

		f[i] = filepath.Join(tmp, name)

		// mark the file for removal as it was generated on the fly
		o.temporaryPaths[f[i]] = true
	}

	o.Paths = append(o.Paths, f...)

	return nil
}

type ComposeStackFiles []string

func (f ComposeStackFiles) applyToComposeStack(o *composeStackOptions) error {
	o.Paths = append(o.Paths, f...)
	return nil
}

type ComposeProfiles []string

func (p ComposeProfiles) applyToComposeStack(o *composeStackOptions) error {
	o.Profiles = append(o.Profiles, p...)
	return nil
}

type StackIdentifier string

func (f StackIdentifier) applyToComposeStack(o *composeStackOptions) error {
	o.Identifier = string(f)
	return nil
}

func (f StackIdentifier) String() string {
	return string(f)
}

const (
	// RemoveImagesAll - remove all images used by the stack
	RemoveImagesAll RemoveImages = iota
	// RemoveImagesLocal - remove only images that don't have a tag
	RemoveImagesLocal
)

type DockerCompose struct {
	// used to synchronize operations
	lock sync.RWMutex

	// name/identifier of the stack that will be started
	// by default a UUID will be used
	name string

	// paths to stack files that will be considered when compiling the final compose project
	configs []string

	// used to remove temporary files that were generated on the fly
	temporaryConfigs map[string]bool

	// used to set logger in DockerContainer
	logger log.Logger

	// wait strategies that are applied per service when starting the stack
	// only one strategy can be added to a service, to use multiple use wait.ForAll(...)
	waitStrategies map[string]wait.Strategy

	// Used to synchronise writes to the containers.
	containersLock sync.Mutex

	// cache for containers that are part of the stack
	// used in ServiceContainer(...) function to avoid calls to the Docker API
	containers map[string]*testcontainers.DockerContainer

	// cache for networks in the compose stack
	networks map[string]*testcontainers.DockerNetwork

	// docker/compose API service instance used to control the compose stack
	composeService api.Service

	// Docker API client used to interact with single container instances and the Docker API e.g. to list containers
	dockerClient client.APIClient

	// options used to compile the compose project
	// e.g. environment settings, ...
	projectOptions []cli.ProjectOptionsFn

	// profiles applied to the compose project after compilation.
	projectProfiles []string

	// compiled compose project
	// can be nil if the stack wasn't started yet
	project *types.Project

	// sessionID is used to identify the reaper session
	sessionID string

	// provider is used to docker operations.
	provider *testcontainers.DockerProvider
}

func (d *DockerCompose) ServiceContainer(ctx context.Context, svcName string) (*testcontainers.DockerContainer, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.lookupContainer(ctx, svcName)
}

func (d *DockerCompose) Services() []string {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.project.ServiceNames()
}

func (d *DockerCompose) Down(ctx context.Context, opts ...StackDownOption) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	options := stackDownOptions{
		DownOptions: api.DownOptions{
			Project: d.project,
		},
	}

	for i := range opts {
		opts[i].applyToStackDown(&options)
	}
	defer func() {
		for cfg := range d.temporaryConfigs {
			_ = os.Remove(cfg)
		}
	}()

	return d.composeService.Down(ctx, d.name, options.DownOptions)
}

func (d *DockerCompose) Up(ctx context.Context, opts ...StackUpOption) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.project, err = d.compileProject(ctx)
	if err != nil {
		return err
	}

	upOptions := stackUpOptions{
		Services:             d.project.ServiceNames(),
		Recreate:             api.RecreateDiverged,
		RecreateDependencies: api.RecreateDiverged,
		Project:              d.project,
	}

	for i := range opts {
		opts[i].applyToStackUp(&upOptions)
	}

	if len(upOptions.Services) != len(d.project.Services) {
		sort.Strings(upOptions.Services)

		filteredServices := types.Services{}

		for _, srv := range upOptions.Services {
			if srvConfig, ok := d.project.Services[srv]; ok {
				filteredServices[srv] = srvConfig
			}
		}

		d.project.Services = filteredServices
	}

	err = d.composeService.Up(ctx, d.project, api.UpOptions{
		Create: api.CreateOptions{
			Build: &api.BuildOptions{
				Services: upOptions.Services,
			},
			Services:             upOptions.Services,
			Recreate:             upOptions.Recreate,
			RecreateDependencies: upOptions.RecreateDependencies,
			RemoveOrphans:        upOptions.RemoveOrphans,
		},
		Start: api.StartOptions{
			Project: upOptions.Project,
			Wait:    upOptions.Wait,
		},
	})
	if err != nil {
		return fmt.Errorf("compose up: %w", err)
	}

	err = d.lookupNetworks(ctx)
	if err != nil {
		return err
	}

	var termSignals []chan bool
	var reaper *testcontainers.Reaper
	if !d.provider.Config().Config.RyukDisabled {
		// NewReaper is deprecated: we need to find a way to create the reaper for compose
		// bypassing the deprecation.
		reaper, err = testcontainers.NewReaper(ctx, testcontainers.SessionID(), d.provider, "")
		if err != nil {
			return fmt.Errorf("create reaper: %w", err)
		}

		// Cleanup on error, otherwise set termSignal to nil before successful return.
		defer func() {
			if len(termSignals) == 0 {
				// Need to call Connect at least once to ensure the initial
				// connection is cleaned up.
				termSignal, errc := reaper.Connect()
				if errc != nil {
					err = errors.Join(err, fmt.Errorf("reaper connect: %w", errc))
				} else {
					termSignal <- true
				}
			}

			if err == nil {
				// No need to cleanup.
				return
			}

			for _, ts := range termSignals {
				ts <- true
			}
		}()

		// Connect to the reaper and set the termination signal for each network.
		for _, n := range d.networks {
			termSignal, err := reaper.Connect()
			if err != nil {
				return fmt.Errorf("reaper connect: %w", err)
			}

			n.SetTerminationSignal(termSignal)
			termSignals = append(termSignals, termSignal)
		}
	}

	errGrpContainers, errGrpCtx := errgroup.WithContext(ctx)

	// Lookup the containers for each service and connect them
	// to the reaper if needed.
	var termSignalsMtx sync.Mutex
	for _, srv := range d.project.Services {
		srv := srv
		errGrpContainers.Go(func() error {
			dc, err := d.lookupContainer(errGrpCtx, srv.Name)
			if err != nil {
				return err
			}

			if reaper != nil {
				termSignal, err := reaper.Connect()
				if err != nil {
					return fmt.Errorf("reaper connect: %w", err)
				}

				dc.SetTerminationSignal(termSignal)

				termSignalsMtx.Lock()
				defer termSignalsMtx.Unlock()
				termSignals = append(termSignals, termSignal)
			}

			return nil
		})
	}

	// wait here for the containers lookup to finish
	if err := errGrpContainers.Wait(); err != nil {
		return err
	}

	if len(d.waitStrategies) == 0 {
		return nil
	}

	errGrpWait, errGrpCtx := errgroup.WithContext(ctx)

	for svc, strategy := range d.waitStrategies { // pinning the variables
		svc := svc
		strategy := strategy

		errGrpWait.Go(func() error {
			target, err := d.lookupContainer(errGrpCtx, svc)
			if err != nil {
				return err
			}

			return strategy.WaitUntilReady(errGrpCtx, target)
		})
	}

	if err := errGrpWait.Wait(); err != nil {
		return fmt.Errorf("wait for services: %w", err)
	}

	return nil
}

func (d *DockerCompose) WaitForService(s string, strategy wait.Strategy) ComposeStack {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.waitStrategies[s] = strategy
	return d
}

func (d *DockerCompose) WithEnv(m map[string]string) ComposeStack {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.projectOptions = append(d.projectOptions, withEnv(m))
	return d
}

func (d *DockerCompose) WithOsEnv() ComposeStack {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.projectOptions = append(d.projectOptions, cli.WithOsEnv)
	return d
}

// cachedContainer returns the cached container for svcName or nil if it doesn't exist.
func (d *DockerCompose) cachedContainer(svcName string) *testcontainers.DockerContainer {
	d.containersLock.Lock()
	defer d.containersLock.Unlock()

	return d.containers[svcName]
}

// lookupContainer is used to retrieve the container instance from the cache or the Docker API.
//
// Safe for concurrent calls.
func (d *DockerCompose) lookupContainer(ctx context.Context, svcName string) (*testcontainers.DockerContainer, error) {
	if c := d.cachedContainer(svcName); c != nil {
		return c, nil
	}

	containers, err := d.dockerClient.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", api.ProjectLabel, d.name)),
			filters.Arg("label", fmt.Sprintf("%s=%s", api.ServiceLabel, svcName)),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("container list: %w", err)
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("no container found for service name %s", svcName)
	}

	ctr, err := d.provider.ContainerFromType(ctx, containers[0])
	if err != nil {
		return nil, fmt.Errorf("container from type: %w", err)
	}

	d.containersLock.Lock()
	defer d.containersLock.Unlock()
	d.containers[svcName] = ctr

	return ctr, nil
}

// lookupNetworks is used to retrieve the networks that are part of the compose stack.
//
// Safe for concurrent calls.
func (d *DockerCompose) lookupNetworks(ctx context.Context) error {
	networks, err := d.dockerClient.NetworkList(ctx, dockernetwork.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", api.ProjectLabel, d.name)),
		),
	})
	if err != nil {
		return fmt.Errorf("network list: %w", err)
	}

	for _, n := range networks {
		dn := &testcontainers.DockerNetwork{
			ID:     n.ID,
			Name:   n.Name,
			Driver: n.Driver,
		}

		d.networks[n.ID] = dn
	}

	return nil
}

func (d *DockerCompose) compileProject(ctx context.Context) (*types.Project, error) {
	const nameAndDefaultConfigPath = 2
	projectOptions := make([]cli.ProjectOptionsFn, len(d.projectOptions), len(d.projectOptions)+nameAndDefaultConfigPath)

	copy(projectOptions, d.projectOptions)
	projectOptions = append(projectOptions, cli.WithName(d.name), cli.WithDefaultConfigPath)

	compiledOptions, err := cli.NewProjectOptions(d.configs, projectOptions...)
	if err != nil {
		return nil, fmt.Errorf("new project options: %w", err)
	}

	proj, err := compiledOptions.LoadProject(ctx)
	if err != nil {
		return nil, fmt.Errorf("load project: %w", err)
	}

	if len(d.projectProfiles) > 0 {
		proj, err = proj.WithProfiles(d.projectProfiles)
		if err != nil {
			return nil, fmt.Errorf("with profiles: %w", err)
		}
	}

	for i, s := range proj.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     proj.Name,
			api.ServiceLabel:     s.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  proj.WorkingDir,
			api.ConfigFilesLabel: strings.Join(proj.ComposeFiles, ","),
			api.OneoffLabel:      "False", // default, will be overridden by `run` command
		}

		testcontainers.AddGenericLabels(s.CustomLabels)

		for i, envFile := range compiledOptions.EnvFiles {
			// add a label for each env file, indexed by its position
			s.CustomLabels[fmt.Sprintf("%s.%d", api.EnvironmentFileLabel, i)] = envFile
		}

		proj.Services[i] = s
	}

	for key, n := range proj.Networks {
		n.Labels = map[string]string{
			api.ProjectLabel: proj.Name,
			api.NetworkLabel: n.Name,
			api.VersionLabel: api.ComposeVersion,
		}

		testcontainers.AddGenericLabels(n.Labels)

		proj.Networks[key] = n
	}

	return proj, nil
}

func withEnv(env map[string]string) func(*cli.ProjectOptions) error {
	return func(options *cli.ProjectOptions) error {
		for k, v := range env {
			if _, ok := options.Environment[k]; ok {
				return fmt.Errorf("environment with key %s already set", k)
			}
			options.Environment[k] = v
		}

		return nil
	}
}

func makeClient(*command.DockerCli) (client.APIClient, error) {
	dockerClient, err := testcontainers.NewDockerClientWithOpts(context.Background())
	if err != nil {
		return nil, fmt.Errorf("new docker client: %w", err)
	}
	return dockerClient, nil
}
