package testcontainers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/testcontainers/testcontainers-go/wait"
	"gopkg.in/yaml.v3"

)

const (
	envProjectName = "COMPOSE_PROJECT_NAME"
	envComposeFile = "COMPOSE_FILE"
)

// DockerCompose defines the contract for running Docker Compose
type DockerCompose interface {
	Down() ExecError
	Invoke() ExecError
	WaitForService(string, wait.Strategy) DockerCompose
	WithCommand([]string) DockerCompose
	WithEnv(map[string]string) DockerCompose
	WithExposedService(string, int, wait.Strategy) DockerCompose
}

type waitService struct {
	service       string
	publishedPort int
}

// LocalDockerCompose represents a Docker Compose execution using local binary
// docker-compose or docker-compose.exe, depending on the underlying platform
type LocalDockerCompose struct {
	*LocalDockerComposeOptions
	ComposeFilePaths     []string
	absComposeFilePaths  []string
	Identifier           string
	Cmd                  []string
	Env                  map[string]string
	Services             map[string]interface{}
	waitStrategySupplied bool
	WaitStrategyMap      map[waitService]wait.Strategy
	Executor             ComposeExecutor
	Context              context.Context
}

type (
	// LocalDockerComposeOptions defines options applicable to LocalDockerCompose
	LocalDockerComposeOptions struct {
		Logger Logging
	}

	// LocalDockerComposeOption defines a common interface to modify LocalDockerComposeOptions
	// These options can be passed to NewLocalDockerCompose in a variadic way to customize the returned LocalDockerCompose instance
	LocalDockerComposeOption interface {
		ApplyToLocalCompose(opts *LocalDockerComposeOptions)
	}

	// LocalDockerComposeOptionsFunc is a shorthand to implement the LocalDockerComposeOption interface
	LocalDockerComposeOptionsFunc func(opts *LocalDockerComposeOptions)
)

func (f LocalDockerComposeOptionsFunc) ApplyToLocalCompose(opts *LocalDockerComposeOptions) {
	f(opts)
}

func NewDockerCompose(
	filePaths []string,
	identifier string,
	executor ComposeExecutor,
	ctx context.Context,
	opts ...LocalDockerComposeOption) *LocalDockerCompose {

	dc := &LocalDockerCompose{
		LocalDockerComposeOptions: &LocalDockerComposeOptions{
			Logger: Logger,
		},
		Executor: executor,
		Context:  ctx,
	}

	for idx := range opts {
		opts[idx].ApplyToLocalCompose(dc.LocalDockerComposeOptions)
	}

	//dc.Executable = "docker-compose"
	//if runtime.GOOS == "windows" {
	//	dc.Executable = "docker-compose.exe"
	//}

	dc.ComposeFilePaths = filePaths

	dc.absComposeFilePaths = make([]string, len(filePaths))
	for i, cfp := range dc.ComposeFilePaths {
		abs, _ := filepath.Abs(cfp)
		dc.absComposeFilePaths[i] = abs
	}

	_ = dc.validate()

	dc.Identifier = strings.ToLower(identifier)
	dc.waitStrategySupplied = false
	dc.WaitStrategyMap = make(map[waitService]wait.Strategy)

	return dc
}

// NewLocalDockerCompose returns an instance of the local Docker Compose, using an
// array of Docker Compose file paths and an identifier for the Compose execution.
//
// It will iterate through the array adding '-f compose-file-path' flags to the local
// Docker Compose execution. The identifier represents the name of the execution,
// which will define the name of the underlying Docker network and the name of the
// running Compose services.
func NewLocalDockerCompose(filePaths []string, identifier string, opts ...LocalDockerComposeOption) *LocalDockerCompose {
	return NewDockerCompose(filePaths, identifier, LocalDockerComposeExecutor, nil, opts...)
}

// Down executes docker-compose down
func (dc *LocalDockerCompose) Down() ExecError {
	return executeCompose(dc, []string{"down", "--remove-orphans", "--volumes"})
}

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

func (dc *LocalDockerCompose) containerNameFromServiceName(service, separator string) string {
	return dc.Identifier + separator + service
}

func (dc *LocalDockerCompose) applyStrategyToRunningContainer() error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	cli.NegotiateAPIVersion(context.Background())

	for k := range dc.WaitStrategyMap {
		containerName := dc.containerNameFromServiceName(k.service, "_")
		composeV2ContainerName := dc.containerNameFromServiceName(k.service, "-")
		f := filters.NewArgs(
			filters.Arg("name", containerName),
			filters.Arg("name", composeV2ContainerName),
			filters.Arg("name", k.service))
		containerListOptions := types.ContainerListOptions{Filters: f}
		containers, err := cli.ContainerList(context.Background(), containerListOptions)
		if err != nil {
			return fmt.Errorf("error %w occured while filtering the service %s: %d by name and published port", err, k.service, k.publishedPort)
		}

		if len(containers) == 0 {
			return fmt.Errorf("service with name %s not found in list of running containers", k.service)
		}

		// The length should always be a list of 1, since we are matching one service name at a time
		if l := len(containers); l > 1 {
			return fmt.Errorf("expecting only one running container for %s but got %d", k.service, l)
		}
		container := containers[0]
		strategy := dc.WaitStrategyMap[k]
		dockerProvider, err := NewDockerProvider(WithLogger(dc.Logger))
		if err != nil {
			return fmt.Errorf("unable to create new Docker Provider: %w", err)
		}
		dockercontainer := &DockerContainer{ID: container.ID, WaitingFor: strategy, provider: dockerProvider, logger: dc.Logger}
		err = strategy.WaitUntilReady(context.Background(), dockercontainer)
		if err != nil {
			return fmt.Errorf("Unable to apply wait strategy %v to service %s due to %w", strategy, k.service, err)
		}
	}
	return nil
}

// Invoke invokes the docker compose
func (dc *LocalDockerCompose) Invoke() ExecError {
	return executeCompose(dc, dc.Cmd)
}

// WaitForService sets the strategy for the service that is to be waited on
func (dc *LocalDockerCompose) WaitForService(service string, strategy wait.Strategy) DockerCompose {
	dc.waitStrategySupplied = true
	dc.WaitStrategyMap[waitService{service: service}] = strategy
	return dc
}

// WithCommand assigns the command
func (dc *LocalDockerCompose) WithCommand(cmd []string) DockerCompose {
	dc.Cmd = cmd
	return dc
}

// WithEnv assigns the environment
func (dc *LocalDockerCompose) WithEnv(env map[string]string) DockerCompose {
	dc.Env = env
	return dc
}

// WithExposedService sets the strategy for the service that is to be waited on. If multiple strategies
// are given for a single service running on different ports, both strategies will be applied on the same container
func (dc *LocalDockerCompose) WithExposedService(service string, port int, strategy wait.Strategy) DockerCompose {
	dc.waitStrategySupplied = true
	dc.WaitStrategyMap[waitService{service: service, publishedPort: port}] = strategy
	return dc
}

// validate checks if the files to be run in the compose are valid YAML files, setting up
// references to all services in them
func (dc *LocalDockerCompose) validate() error {
	type compose struct {
		Services map[string]interface{}
	}

	for _, abs := range dc.absComposeFilePaths {
		c := compose{}

		yamlFile, err := ioutil.ReadFile(abs)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(yamlFile, &c)
		if err != nil {
			return err
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
	dirContext string, environment map[string]string, binary string, args []string) ExecError {

	var errStdout, errStderr error

	cmd := exec.Command(binary, args...)
	cmd.Dir = dirContext
	cmd.Env = os.Environ()

	for key, value := range environment {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	stdout := newCapturingPassThroughWriter(os.Stdout)
	stderr := newCapturingPassThroughWriter(os.Stderr)

	err := cmd.Start()
	if err != nil {
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

func executeCompose(dc *LocalDockerCompose, args []string) ExecError {
	environment := dc.getDockerComposeEnvironment()
	for k, v := range dc.Env {
		environment[k] = v
	}

	pwd, err := filepath.Abs(".")
	if err != nil {
		return ExecError{
			Command: []string{},
			Error:   err,
		}
	}

	if len(dc.absComposeFilePaths) > 0 {
		pwd, _ = filepath.Split(dc.absComposeFilePaths[0])
	}

	execErr := dc.Executor(ComposeExecutorOptions{
		Env:          environment,
		Pwd:          pwd,
		Args:         args,
		ComposeFiles: dc.absComposeFilePaths,
		Context:      dc.Context,
	})

	//execErr := execute(pwd, environment, dc.Executable, cmds)
	err = execErr.Error
	if err != nil {
		args := strings.Join(dc.Cmd, " ")
		return ExecError{
			Command: []string{},
			Error:   fmt.Errorf("Local Docker compose exited abnormally whilst running: [%v]. %s", args, err.Error()),
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
	_, err := exec.LookPath(binary)

	return err
}
