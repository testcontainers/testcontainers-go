package compose

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"gopkg.in/yaml.v3"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	_ ComposeVersion = (*composeVersion1)(nil)
	_ ComposeVersion = (*composeVersion2)(nil)
)

type ComposeVersion interface {
	Format(parts ...string) string
}

type composeVersion1 struct{}

func (c composeVersion1) Format(parts ...string) string {
	return strings.Join(parts, "_")
}

type composeVersion2 struct{}

func (c composeVersion2) Format(parts ...string) string {
	return strings.Join(parts, "-")
}

// Deprecated: use ComposeStack instead
// LocalDockerCompose represents a Docker Compose execution using local binary
// docker compose or docker.exe compose, depending on the underlying platform
type LocalDockerCompose struct {
	ComposeVersion
	*LocalDockerComposeOptions
	Executable           string
	composeSubcommand    string
	ComposeFilePaths     []string
	absComposeFilePaths  []string
	Identifier           string
	Cmd                  []string
	Env                  map[string]string
	Services             map[string]any
	waitStrategySupplied bool
	WaitStrategyMap      map[waitService]wait.Strategy
}

type (
	// Deprecated: it will be removed in the next major release
	// LocalDockerComposeOptions defines options applicable to LocalDockerCompose
	LocalDockerComposeOptions struct {
		Logger log.Logger
	}

	// Deprecated: it will be removed in the next major release
	// LocalDockerComposeOption defines a common interface to modify LocalDockerComposeOptions
	// These options can be passed to NewLocalDockerCompose in a variadic way to customize the returned LocalDockerCompose instance
	LocalDockerComposeOption interface {
		ApplyToLocalCompose(opts *LocalDockerComposeOptions)
	}

	// Deprecated: it will be removed in the next major release
	// LocalDockerComposeOptionsFunc is a shorthand to implement the LocalDockerComposeOption interface
	LocalDockerComposeOptionsFunc func(opts *LocalDockerComposeOptions)
)

type ComposeLoggerOption struct {
	logger log.Logger
}

// WithLogger is a generic option that implements LocalDockerComposeOption
// It replaces the global Logging implementation with a user defined one e.g. to aggregate logs from testcontainers
// with the logs of specific test case
func WithLogger(logger log.Logger) ComposeLoggerOption {
	return ComposeLoggerOption{
		logger: logger,
	}
}

// Deprecated: it will be removed in the next major release
func (o ComposeLoggerOption) ApplyToLocalCompose(opts *LocalDockerComposeOptions) {
	opts.Logger = o.logger
}

func (o ComposeLoggerOption) applyToComposeStack(opts *composeStackOptions) error {
	opts.Logger = o.logger
	return nil
}

// Deprecated: it will be removed in the next major release
func (f LocalDockerComposeOptionsFunc) ApplyToLocalCompose(opts *LocalDockerComposeOptions) {
	f(opts)
}

// Deprecated: it will be removed in the next major release
// Down executes docker compose down
func (dc *LocalDockerCompose) Down() ExecError {
	return executeCompose(dc, []string{"down", "--remove-orphans", "--volumes"})
}

// Deprecated: it will be removed in the next major release
func (dc *LocalDockerCompose) getDockerComposeEnvironment() map[string]string {
	environment := map[string]string{}

	composeFileEnvVariableValue := ""
	for _, abs := range dc.absComposeFilePaths {
		composeFileEnvVariableValue += abs + string(os.PathListSeparator)
	}

	environment[envProjectName] = dc.Identifier
	environment[envComposeFile] = composeFileEnvVariableValue

	return environment
}

// Deprecated: it will be removed in the next major release
func (dc *LocalDockerCompose) containerNameFromServiceName(service, separator string) string {
	return dc.Identifier + separator + service
}

// Deprecated: it will be removed in the next major release
func (dc *LocalDockerCompose) applyStrategyToRunningContainer() error {
	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	if err != nil {
		return fmt.Errorf("new docker client: %w", err)
	}
	defer cli.Close()

	for k := range dc.WaitStrategyMap {
		containerName := dc.containerNameFromServiceName(k.service, "_")
		composeV2ContainerName := dc.containerNameFromServiceName(k.service, "-")
		f := filters.NewArgs(
			filters.Arg("name", containerName),
			filters.Arg("name", composeV2ContainerName),
			filters.Arg("name", k.service))
		containerListOptions := container.ListOptions{Filters: f, All: true}
		containers, err := cli.ContainerList(context.Background(), containerListOptions)
		if err != nil {
			return fmt.Errorf("container list service %q: %w", k.service, err)
		}

		if len(containers) == 0 {
			return fmt.Errorf("service with name %q not found in list of running containers", k.service)
		}

		// The length should always be a list of 1, since we are matching one service name at a time
		if l := len(containers); l > 1 {
			return fmt.Errorf("expecting only one running container for %q but got %d", k.service, l)
		}
		container := containers[0]
		strategy := dc.WaitStrategyMap[k]
		dockerProvider, err := testcontainers.NewDockerProvider(testcontainers.WithLogger(dc.Logger))
		if err != nil {
			return fmt.Errorf("new docker provider: %w", err)
		}
		defer dockerProvider.Close()

		dockercontainer := &testcontainers.DockerContainer{ID: container.ID, WaitingFor: strategy}
		dockercontainer.SetLogger(dc.Logger)
		dockercontainer.SetProvider(dockerProvider)

		err = strategy.WaitUntilReady(context.Background(), dockercontainer)
		if err != nil {
			return fmt.Errorf("wait until ready %v to service %q due: %w", strategy, k.service, err)
		}
	}
	return nil
}

// Deprecated: it will be removed in the next major release
// Invoke invokes the docker compose
func (dc *LocalDockerCompose) Invoke() ExecError {
	return executeCompose(dc, dc.Cmd)
}

// Deprecated: it will be removed in the next major release
// WaitForService sets the strategy for the service that is to be waited on
func (dc *LocalDockerCompose) WaitForService(service string, strategy wait.Strategy) DockerComposer {
	dc.waitStrategySupplied = true
	dc.WaitStrategyMap[waitService{service: service}] = strategy
	return dc
}

// Deprecated: it will be removed in the next major release
// WithCommand assigns the command
func (dc *LocalDockerCompose) WithCommand(cmd []string) DockerComposer {
	dc.Cmd = cmd
	return dc
}

// Deprecated: it will be removed in the next major release
// WithEnv assigns the environment
func (dc *LocalDockerCompose) WithEnv(env map[string]string) DockerComposer {
	dc.Env = env
	return dc
}

// Deprecated: it will be removed in the next major release
// WithExposedService sets the strategy for the service that is to be waited on. If multiple strategies
// are given for a single service running on different ports, both strategies will be applied on the same container
func (dc *LocalDockerCompose) WithExposedService(service string, port int, strategy wait.Strategy) DockerComposer {
	dc.waitStrategySupplied = true
	dc.WaitStrategyMap[waitService{service: service, publishedPort: port}] = strategy
	return dc
}

// Deprecated: it will be removed in the next major release
// determineVersion checks which version of docker compose is installed
// depending on the version services names are composed in a different way
func (dc *LocalDockerCompose) determineVersion() error {
	execErr := executeCompose(dc, []string{"version", "--short"})
	if err := execErr.Error; err != nil {
		return err
	}

	components := bytes.Split(execErr.StdoutOutput, []byte("."))
	if componentsLen := len(components); componentsLen < 3 {
		return fmt.Errorf("expected +3 version components in %s", execErr.StdoutOutput)
	}

	majorVersion, err := strconv.ParseInt(string(components[0]), 10, 8)
	if err != nil {
		return fmt.Errorf("parsing major version: %w", err)
	}

	switch majorVersion {
	case 1:
		dc.ComposeVersion = composeVersion1{}
	case 2:
		dc.ComposeVersion = composeVersion2{}
	default:
		return fmt.Errorf("unexpected compose version %d", majorVersion)
	}

	return nil
}

// Deprecated: it will be removed in the next major release
// validate checks if the files to be run in the compose are valid YAML files, setting up
// references to all services in them
func (dc *LocalDockerCompose) validate() error {
	type compose struct {
		Services map[string]any
	}

	for _, abs := range dc.absComposeFilePaths {
		c := compose{}

		yamlFile, err := os.ReadFile(abs)
		if err != nil {
			return fmt.Errorf("read compose file %q: %w", abs, err)
		}
		err = yaml.Unmarshal(yamlFile, &c)
		if err != nil {
			return fmt.Errorf("unmarshalling file %q: %w", abs, err)
		}

		if dc.Services == nil {
			dc.Services = c.Services
		} else {
			for k, v := range c.Services {
				dc.Services[k] = v
			}
		}
	}

	return nil
}

// ExecError is super struct that holds any information about an execution error, so the client code
// can handle the result
type ExecError struct {
	Command      []string
	StdoutOutput []byte
	StderrOutput []byte
	Error        error
	Stdout       error
	Stderr       error
}

// execute executes a program with arguments and environment variables inside a specific directory
func execute(
	dirContext string, environment map[string]string, binary string, args []string,
) ExecError {
	var errStdout, errStderr error

	cmd := exec.Command(binary, args...)
	cmd.Dir = dirContext
	cmd.Env = os.Environ()

	for key, value := range environment {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	stdoutIn, err := cmd.StdoutPipe()
	if err != nil {
		return ExecError{
			Command: cmd.Args,
			Error:   fmt.Errorf("stdout: %w", err),
		}
	}

	stderrIn, err := cmd.StderrPipe()
	if err != nil {
		return ExecError{
			Command: cmd.Args,
			Error:   fmt.Errorf("stderr: %w", err),
		}
	}

	stdout := newCapturingPassThroughWriter(os.Stdout)
	stderr := newCapturingPassThroughWriter(os.Stderr)

	if err = cmd.Start(); err != nil {
		execCmd := []string{"Starting command", dirContext, binary}
		execCmd = append(execCmd, args...)

		return ExecError{
			// add information about the CMD and arguments used
			Command:      execCmd,
			StdoutOutput: stdout.Bytes(),
			StderrOutput: stderr.Bytes(),
			Error:        err,
			Stderr:       errStderr,
			Stdout:       errStdout,
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()

	execCmd := []string{"Reading std", dirContext, binary}
	execCmd = append(execCmd, args...)

	return ExecError{
		Command:      execCmd,
		StdoutOutput: stdout.Bytes(),
		StderrOutput: stderr.Bytes(),
		Error:        err,
		Stderr:       errStderr,
		Stdout:       errStdout,
	}
}

// Deprecated: it will be removed in the next major release
func executeCompose(dc *LocalDockerCompose, args []string) ExecError {
	if which(dc.Executable) != nil {
		return ExecError{
			Command: []string{dc.Executable},
			Error:   fmt.Errorf("local Docker not found. Is %s on the PATH?", dc.Executable),
		}
	}

	environment := dc.getDockerComposeEnvironment()
	for k, v := range dc.Env {
		environment[k] = v
	}

	// initialise the command with the compose subcommand
	cmds := []string{dc.composeSubcommand}
	pwd := "."
	if len(dc.absComposeFilePaths) > 0 {
		pwd, _ = filepath.Split(dc.absComposeFilePaths[0])

		for _, abs := range dc.absComposeFilePaths {
			cmds = append(cmds, "-f", abs)
		}
	} else {
		cmds = append(cmds, "-f", "docker-compose.yml")
	}
	cmds = append(cmds, args...)

	execErr := execute(pwd, environment, dc.Executable, cmds)
	err := execErr.Error
	if err != nil {
		args := strings.Join(dc.Cmd, " ")
		return ExecError{
			Command: []string{dc.Executable, args},
			Error:   fmt.Errorf("local Docker compose exited abnormally whilst running %s: [%v]. %s", dc.Executable, args, err.Error()),
		}
	}

	if dc.waitStrategySupplied {
		// If the wait strategy has been executed once for all services during startup , disable it so that it is not invoked while tearing down
		dc.waitStrategySupplied = false
		if err := dc.applyStrategyToRunningContainer(); err != nil {
			return ExecError{
				Error: fmt.Errorf("one or more wait strategies could not be applied to the running containers: %w", err),
			}
		}
	}

	return execErr
}

// capturingPassThroughWriter is a writer that remembers
// data written to it and passes it to w
type capturingPassThroughWriter struct {
	buf bytes.Buffer
	w   io.Writer
}

// newCapturingPassThroughWriter creates new capturingPassThroughWriter
func newCapturingPassThroughWriter(w io.Writer) *capturingPassThroughWriter {
	return &capturingPassThroughWriter{
		w: w,
	}
}

func (w *capturingPassThroughWriter) Write(d []byte) (int, error) {
	w.buf.Write(d)
	return w.w.Write(d)
}

// Bytes returns bytes written to the writer
func (w *capturingPassThroughWriter) Bytes() []byte {
	b := w.buf.Bytes()
	if b == nil {
		b = []byte{}
	}
	return b
}

// Which checks if a binary is present in PATH
func which(binary string) error {
	if _, err := exec.LookPath(binary); err != nil {
		return fmt.Errorf("lookup: %w", err)
	}

	return nil
}
