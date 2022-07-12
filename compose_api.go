package testcontainers

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/pkg/api"
	types2 "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"golang.org/x/sync/errgroup"

	"github.com/testcontainers/testcontainers-go/wait"
)

type stackUpOptionFunc func(s *stackUpOptions)

func (f stackUpOptionFunc) applyToStackUp(o *stackUpOptions) {
	f(o)
}

type stackDownOptionFunc func(do *api.DownOptions)

func (f stackDownOptionFunc) applyToStackDown(do *api.DownOptions) {
	f(do)
}

// RunServices is comparable to 'docker-compose run' as it only creates a subset of containers
// instead of all services defined by the project
func RunServices(serviceNames ...string) StackUpOption {
	return stackUpOptionFunc(func(o *stackUpOptions) {
		o.Services = serviceNames
	})
}

// IgnoreOrphans - Ignore legacy containers for services that are not defined in the project
type IgnoreOrphans bool

func (io IgnoreOrphans) applyToStackUp(co *api.CreateOptions, _ *api.StartOptions) {
	co.IgnoreOrphans = bool(io)
}

// RemoveOrphans will cleanup containers that are not declared on the compose model but own the same labels
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

const (
	// RemoveImagesAll - remove all images used by the stack
	RemoveImagesAll RemoveImages = iota
	// RemoveImagesLocal - remove only images that don't have a tag
	RemoveImagesLocal
)

type dockerComposeAPI struct {
	lock           sync.RWMutex
	name           string
	configs        []string
	waitStrategies map[string]wait.Strategy
	containers     map[string]*DockerContainer
	composeService api.Service
	dockerClient   client.APIClient
	projectOptions []cli.ProjectOptionsFn
	project        *types.Project
}

func (d *dockerComposeAPI) Services() []string {
	return d.project.ServiceNames()
}

func (d *dockerComposeAPI) Down(ctx context.Context, opts ...StackDownOption) error {
	options := stackDownOptions{
		DownOptions: api.DownOptions{
			Project: d.project,
		},
	}

	for i := range opts {
		opts[i].applyToStackDown(&options)
	}

	return d.composeService.Down(ctx, d.name, options.DownOptions)
}

func (d *dockerComposeAPI) Up(ctx context.Context, opts ...StackUpOption) (err error) {
	d.project, err = d.compileProject()
	if err != nil {
		return err
	}

	upOptions := stackUpOptions{
		CreateOptions: api.CreateOptions{
			Services:             d.project.ServiceNames(),
			Recreate:             api.RecreateDiverged,
			RecreateDependencies: api.RecreateDiverged,
		},
		StartOptions: api.StartOptions{
			Project: d.project,
		},
	}

	for i := range opts {
		opts[i].applyToStackUp(&upOptions)
	}

	if len(upOptions.Services) != len(d.project.Services) {
		sort.Strings(upOptions.Services)

		filteredServices := make(types.Services, 0, len(d.project.Services))

		for i := range d.project.Services {
			if idx := sort.SearchStrings(upOptions.Services, d.project.Services[i].Name); idx < len(upOptions.Services) && upOptions.Services[idx] == d.project.Services[i].Name {
				filteredServices = append(filteredServices, d.project.Services[i])
			}
		}

		d.project.Services = filteredServices
	}

	err = d.composeService.Up(ctx, d.project, api.UpOptions{
		Create: upOptions.CreateOptions,
		Start:  upOptions.StartOptions,
	})

	if len(d.waitStrategies) == 0 || err != nil {
		return err
	}

	errGrp, errGrpCtx := errgroup.WithContext(ctx)

	for svc, strategy := range d.waitStrategies { // pinning the variables
		svc := svc
		strategy := strategy

		errGrp.Go(func() error {
			target, err := d.ServiceContainer(errGrpCtx, svc)
			if err != nil {
				return err
			}
			return strategy.WaitUntilReady(errGrpCtx, target)
		})
	}

	return errGrp.Wait()
}

func (d *dockerComposeAPI) WaitForService(s string, strategy wait.Strategy) ComposeStack {
	d.waitStrategies[s] = strategy
	return d
}

func (d *dockerComposeAPI) WithEnv(m map[string]string) ComposeStack {
	d.projectOptions = append(d.projectOptions, withEnv(m))
	return d
}

func (d *dockerComposeAPI) ServiceContainer(ctx context.Context, svcName string) (*DockerContainer, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if container, ok := d.containers[svcName]; ok {
		return container, nil
	}

	listOptions := types2.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", api.ProjectLabel, d.name)),
			filters.Arg("label", fmt.Sprintf("%s=%s", api.ServiceLabel, svcName)),
		),
	}

	containers, err := d.dockerClient.ContainerList(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("no container found for service name %s", svcName)
	}

	containerInstance := containers[0]
	container := &DockerContainer{
		ID: containerInstance.ID,
		provider: &DockerProvider{
			client: d.dockerClient,
		},
	}

	d.containers[svcName] = container

	return container, nil
}

func (d *dockerComposeAPI) compileProject() (*types.Project, error) {
	projectOptions := make([]cli.ProjectOptionsFn, len(d.projectOptions), len(d.projectOptions)+2)
	copy(projectOptions, d.projectOptions)
	projectOptions = append(projectOptions, cli.WithName(d.name), cli.WithDefaultConfigPath)

	compiledOptions, err := cli.NewProjectOptions(d.configs, projectOptions...)
	if err != nil {
		return nil, err
	}

	proj, err := cli.ProjectFromOptions(compiledOptions)
	if err != nil {
		return nil, err
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
		if compiledOptions.EnvFile != "" {
			s.CustomLabels[api.EnvironmentFileLabel] = compiledOptions.EnvFile
		}
		proj.Services[i] = s
	}

	return proj, nil
}

func withEnv(env map[string]string) func(*cli.ProjectOptions) error {
	return func(options *cli.ProjectOptions) error {
		for k, v := range env {
			if _, ok := options.Environment[k]; ok {
				return fmt.Errorf("environment with key %s already set", k)
			} else {
				options.Environment[k] = v
			}
		}

		return nil
	}
}
