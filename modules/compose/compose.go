package compose

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/google/uuid"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	envProjectName = "COMPOSE_PROJECT_NAME"
	envComposeFile = "COMPOSE_FILE"
)

var ErrNoStackConfigured = errors.New("no stack files configured")

type composeStackOptions struct {
	Identifier     string
	Paths          []string
	temporaryPaths map[string]bool
	Logger         log.Logger
	Profiles       []string
}

type ComposeStackOption interface {
	applyToComposeStack(o *composeStackOptions) error
}

type stackUpOptions struct {
	// Services defines the services user interacts with
	Services []string
	// Remove legacy containers for services that are not defined in the project
	RemoveOrphans bool
	// Wait won't return until containers reached the running|healthy state
	Wait bool
	// Recreate define the strategy to apply on existing containers
	Recreate string
	// RecreateDependencies define the strategy to apply on dependencies services
	RecreateDependencies string
	// Project is the compose project used to define this app. Might be nil if user ran command just with project name
	Project *types.Project
}

type StackUpOption interface {
	applyToStackUp(o *stackUpOptions)
}

type stackDownOptions struct {
	api.DownOptions
}

type StackDownOption interface {
	applyToStackDown(do *stackDownOptions)
}

// ComposeStack defines operations that can be applied to a parsed compose stack
type ComposeStack interface {
	Up(ctx context.Context, opts ...StackUpOption) error
	Down(ctx context.Context, opts ...StackDownOption) error
	Services() []string
	WaitForService(s string, strategy wait.Strategy) ComposeStack
	WithEnv(m map[string]string) ComposeStack
	WithOsEnv() ComposeStack
	ServiceContainer(ctx context.Context, svcName string) (*testcontainers.DockerContainer, error)
}

// Deprecated: DockerComposer is the old shell escape based API
// use ComposeStack instead
// DockerComposer defines the contract for running Docker Compose
type DockerComposer interface {
	Down() ExecError
	Invoke() ExecError
	WaitForService(string, wait.Strategy) DockerComposer
	WithCommand([]string) DockerComposer
	WithEnv(map[string]string) DockerComposer
	WithExposedService(string, int, wait.Strategy) DockerComposer
}

type waitService struct {
	service       string
	publishedPort int
}

// WithRecreate defines the strategy to apply on existing containers. If any other value than
// api.RecreateNever, api.RecreateForce or api.RecreateDiverged is provided, the default value
// api.RecreateForce will be used.
func WithRecreate(recreate string) StackUpOption {
	return Recreate(recreate)
}

// WithRecreateDependencies defines the strategy to apply on container dependencies. If any other value than
// api.RecreateNever, api.RecreateForce or api.RecreateDiverged is provided, the default value
// api.RecreateForce will be used.
func WithRecreateDependencies(recreate string) StackUpOption {
	return RecreateDependencies(recreate)
}

func WithStackFiles(filePaths ...string) ComposeStackOption {
	return ComposeStackFiles(filePaths)
}

// WithStackReaders supports reading the compose file/s from a reader.
func WithStackReaders(readers ...io.Reader) ComposeStackOption {
	return ComposeStackReaders(readers)
}

// WithProfiles allows to enable/disable services based on the profiles defined in the compose file.
func WithProfiles(profiles ...string) ComposeStackOption {
	return ComposeProfiles(profiles)
}

func NewDockerCompose(filePaths ...string) (*DockerCompose, error) {
	return NewDockerComposeWith(WithStackFiles(filePaths...))
}

func NewDockerComposeWith(opts ...ComposeStackOption) (*DockerCompose, error) {
	composeOptions := composeStackOptions{
		Identifier:     uuid.New().String(),
		temporaryPaths: make(map[string]bool),
		Logger:         log.Default(),
		Profiles:       nil,
	}

	for i := range opts {
		if err := opts[i].applyToComposeStack(&composeOptions); err != nil {
			return nil, fmt.Errorf("apply compose stack option: %w", err)
		}
	}

	if len(composeOptions.Paths) < 1 {
		return nil, ErrNoStackConfigured
	}

	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("new docker client: %w", err)
	}

	if err = dockerCli.Initialize(flags.NewClientOptions(), command.WithInitializeClient(makeClient)); err != nil {
		return nil, fmt.Errorf("initialize docker client: %w", err)
	}

	provider, err := testcontainers.NewDockerProvider(testcontainers.WithLogger(composeOptions.Logger))
	if err != nil {
		return nil, fmt.Errorf("new docker provider: %w", err)
	}

	dockerClient := dockerCli.Client()
	provider.SetClient(dockerClient)

	composeAPI := &DockerCompose{
		name:             composeOptions.Identifier,
		configs:          composeOptions.Paths,
		temporaryConfigs: composeOptions.temporaryPaths,
		logger:           composeOptions.Logger,
		projectProfiles:  composeOptions.Profiles,
		composeService:   compose.NewComposeService(dockerCli),
		dockerClient:     dockerClient,
		waitStrategies:   make(map[string]wait.Strategy),
		containers:       make(map[string]*testcontainers.DockerContainer),
		networks:         make(map[string]*testcontainers.DockerNetwork),
		sessionID:        testcontainers.SessionID(),
		provider:         provider,
	}

	return composeAPI, nil
}

// Deprecated: NewLocalDockerCompose returns a DockerComposer compatible instance which is superseded
// by ComposeStack use NewDockerCompose instead to get a ComposeStack compatible instance
//
// NewLocalDockerCompose returns an instance of the local Docker Compose, using an
// array of Docker Compose file paths and an identifier for the Compose execution.
//
// It will iterate through the array adding '-f compose-file-path' flags to the local
// Docker Compose execution. The identifier represents the name of the execution,
// which will define the name of the underlying Docker network and the name of the
// running Compose services.
func NewLocalDockerCompose(filePaths []string, identifier string, opts ...LocalDockerComposeOption) *LocalDockerCompose {
	dc := &LocalDockerCompose{
		LocalDockerComposeOptions: &LocalDockerComposeOptions{
			Logger: log.Default(),
		},
	}

	for idx := range opts {
		opts[idx].ApplyToLocalCompose(dc.LocalDockerComposeOptions)
	}

	dc.Executable = "docker"
	if runtime.GOOS == "windows" {
		dc.Executable = "docker.exe"
	}

	dc.composeSubcommand = "compose"
	dc.ComposeFilePaths = filePaths

	dc.absComposeFilePaths = make([]string, len(filePaths))
	for i, cfp := range dc.ComposeFilePaths {
		abs, _ := filepath.Abs(cfp)
		dc.absComposeFilePaths[i] = abs
	}

	_ = dc.determineVersion()
	_ = dc.validate()

	dc.Identifier = strings.ToLower(identifier)
	dc.waitStrategySupplied = false
	dc.WaitStrategyMap = make(map[waitService]wait.Strategy)

	return dc
}
