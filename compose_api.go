package testcontainers

import (
	"context"
	"fmt"
	"strings"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	types2 "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"golang.org/x/sync/errgroup"

	"github.com/testcontainers/testcontainers-go/wait"
)

func NewDockerComposeApi(filePaths []string, identifier string, opts ...LocalDockerComposeOption) (*dockerComposeApi, error) {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, err
	}

	if err = dockerCli.Initialize(&flags.ClientOptions{
		Common: new(flags.CommonOptions),
	}); err != nil {
		return nil, err
	}

	composeApi := &dockerComposeApi{
		name:            identifier,
		configs:         filePaths,
		composeService:  compose.NewComposeService(dockerCli),
		dockerClient:    dockerCli.Client(),
		waitStrategyMap: make(map[waitService]wait.Strategy),
	}

	return composeApi, nil
}

type dockerComposeApi struct {
	name            string
	configs         []string
	waitStrategyMap map[waitService]wait.Strategy
	composeService  api.Service
	dockerClient    client.APIClient
	projectOptions  []cli.ProjectOptionsFn
	project         *types.Project
}

func (d *dockerComposeApi) ServiceNames() []string {
	return d.project.ServiceNames()
}

func (d *dockerComposeApi) Services() types.Services {
	return d.project.AllServices()
}

func (d *dockerComposeApi) Down(ctx context.Context) error {
	return d.composeService.Down(ctx, d.name, api.DownOptions{})
}

func (d *dockerComposeApi) Up(ctx context.Context) error {
	projectOptions := make([]cli.ProjectOptionsFn, len(d.projectOptions), len(d.projectOptions)+2)
	copy(projectOptions, d.projectOptions)
	projectOptions = append(projectOptions, cli.WithName(d.name), cli.WithDefaultConfigPath)

	opts, err := cli.NewProjectOptions(d.configs, projectOptions...)
	if err != nil {
		return err
	}

	d.project, err = cli.ProjectFromOptions(opts)
	if err != nil {
		return err
	}

	ensureDefaultValues(d.project, opts)

	err = d.composeService.Up(ctx, d.project, api.UpOptions{
		Create: api.CreateOptions{
			Services:             d.project.ServiceNames(),
			RemoveOrphans:        true,
			IgnoreOrphans:        false,
			Recreate:             api.RecreateDiverged,
			RecreateDependencies: api.RecreateDiverged,
		},
		Start: api.StartOptions{
			Project:  d.project,
			AttachTo: d.project.ServiceNames(),
			Wait:     true,
		},
	})

	if len(d.waitStrategyMap) == 0 || err != nil {
		return err
	}

	grp, grpCtx := errgroup.WithContext(ctx)

	for svc, strategy := range d.waitStrategyMap { // pinning the variables
		svc := svc
		strategy := strategy

		grp.Go(func() error {
			target, err := d.serviceContainer(grpCtx, svc.service)
			if err != nil {
				return err
			}
			return strategy.WaitUntilReady(grpCtx, target)
		})
	}

	return grp.Wait()
}

func (d *dockerComposeApi) WaitForService(s string, strategy wait.Strategy) *dockerComposeApi {
	d.waitStrategyMap[waitService{service: s}] = strategy
	return d
}

func (d *dockerComposeApi) WithEnv(m map[string]string) *dockerComposeApi {
	d.projectOptions = append(d.projectOptions, withEnv(m))
	return d
}

func (d *dockerComposeApi) WithExposedService(s string, i int, strategy wait.Strategy) *dockerComposeApi {
	d.waitStrategyMap[waitService{service: s, publishedPort: i}] = strategy
	return d
}

func (d *dockerComposeApi) serviceContainer(ctx context.Context, svcName string) (*DockerContainer, error) {
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

	container := containers[0]
	return &DockerContainer{
		ID: container.ID,
		provider: &DockerProvider{
			client: d.dockerClient,
		},
	}, nil
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

func ensureDefaultValues(proj *types.Project, opts *cli.ProjectOptions) {
	for i, s := range proj.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     proj.Name,
			api.ServiceLabel:     s.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  proj.WorkingDir,
			api.ConfigFilesLabel: strings.Join(proj.ComposeFiles, ","),
			api.OneoffLabel:      "False", // default, will be overridden by `run` command
		}
		if opts.EnvFile != "" {
			s.CustomLabels[api.EnvironmentFileLabel] = opts.EnvFile
		}
		proj.Services[i] = s
	}
}
