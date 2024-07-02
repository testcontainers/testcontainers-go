package compose

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/google/uuid"

	"github.com/testcontainers/testcontainers-go"
	tcconfig "github.com/testcontainers/testcontainers-go/internal/config"
	tclog "github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

var ErrNoStackConfigured = errors.New("no stack files configured")

type composeStackOptions struct {
	Identifier     string
	Paths          []string
	temporaryPaths map[string]bool
	Logger         tclog.Logging
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

func NewDockerCompose(filePaths ...string) (*dockerCompose, error) {
	return NewDockerComposeWith(WithStackFiles(filePaths...))
}

func NewDockerComposeWith(opts ...ComposeStackOption) (*dockerCompose, error) {
	composeOptions := composeStackOptions{
		Identifier:     uuid.New().String(),
		temporaryPaths: make(map[string]bool),
		Logger:         tclog.StandardLogger(),
	}

	for i := range opts {
		if err := opts[i].applyToComposeStack(&composeOptions); err != nil {
			return nil, err
		}
	}

	if len(composeOptions.Paths) < 1 {
		return nil, ErrNoStackConfigured
	}

	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, err
	}

	if err = dockerCli.Initialize(flags.NewClientOptions(), command.WithInitializeClient(makeClient)); err != nil {
		return nil, err
	}

	var composeReaper *testcontainers.Reaper
	if !tcconfig.Read().RyukDisabled {
		// Initialise the reaper for the compose module
		r, err := testcontainers.NewReaper(context.Background(), testcontainers.SessionID())
		if err != nil {
			return nil, fmt.Errorf("failed to create reaper for compose: %w", err)
		}

		composeReaper = r
	}

	composeAPI := &dockerCompose{
		name:             composeOptions.Identifier,
		configs:          composeOptions.Paths,
		temporaryConfigs: composeOptions.temporaryPaths,
		logger:           composeOptions.Logger,
		composeService:   compose.NewComposeService(dockerCli),
		dockerClient:     dockerCli.Client(),
		waitStrategies:   make(map[string]wait.Strategy),
		containers:       make(map[string]*testcontainers.DockerContainer),
		networks:         make(map[string]*testcontainers.DockerNetwork),
		sessionID:        testcontainers.SessionID(),
		reaper:           composeReaper,
	}

	return composeAPI, nil
}
