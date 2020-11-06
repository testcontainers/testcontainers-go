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
	"runtime"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/testcontainers/testcontainers-go/wait"
	"gopkg.in/yaml.v2"
)

const (
	envProjectName = "COMPOSE_PROJECT_NAME"
	envComposeFile = "COMPOSE_FILE"
)

// DockerCompose defines the contract for running Docker Compose
type DockerCompose interface {
	Down() ExecError
	Invoke() ExecError
	WithCommand([]string) DockerCompose
	WithEnv(map[string]string) DockerCompose
	WithExposedService(string, wait.Strategy) DockerCompose
}

// LocalDockerCompose represents a Docker Compose execution using local binary
// docker-compose or docker-compose.exe, depending on the underlying platform
type LocalDockerCompose struct {
	Executable           string
	ComposeFilePaths     []string
	absComposeFilePaths  []string
	Identifier           string
	Cmd                  []string
	Env                  map[string]string
	Services             map[string]interface{}
	waitStrategySupplied bool
	WaitStrategyMap      map[string]wait.Strategy
}

// NewLocalDockerCompose returns an instance of the local Docker Compose, using an
// array of Docker Compose file paths and an identifier for the Compose execution.
//
// It will iterate through the array adding '-f compose-file-path' flags to the local
// Docker Compose execution. The identifier represents the name of the execution,
// which will define the name of the underlying Docker network and the name of the
// running Compose services.
func NewLocalDockerCompose(filePaths []string, identifier string) *LocalDockerCompose {
	dc := &LocalDockerCompose{}

	dc.Executable = "docker-compose"
	if runtime.GOOS == "windows" {
		dc.Executable = "docker-compose.exe"
	}

	dc.ComposeFilePaths = filePaths

	dc.absComposeFilePaths = make([]string, len(filePaths))
	for i, cfp := range dc.ComposeFilePaths {
		abs, _ := filepath.Abs(cfp)
		dc.absComposeFilePaths[i] = abs
	}

	dc.validate()

	dc.Identifier = strings.ToLower(identifier)
	dc.waitStrategySupplied = false
	dc.WaitStrategyMap = make(map[string]wait.Strategy)

	return dc
}

// Down executes docker-compose down
func (dc *LocalDockerCompose) Down() ExecError {
	return executeCompose(dc, []string{"down", "--remove-orphans"})
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

func (dc *LocalDockerCompose) applyStrategyToRunningContainer() error {

	failedStrategies := make(map[string]wait.Strategy)

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	for k := range dc.WaitStrategyMap {
		// Docker compose appends "_1" to every started service by default. Trimming to match running docker container
		f := filters.NewArgs(filters.Arg("name", strings.TrimSuffix(k, "_1")))
		containerListOptions := types.ContainerListOptions{Filters: f}
		containers, cErr := cli.ContainerList(context.Background(), containerListOptions)
		if cErr != nil {
			return cErr
		}
		if len(containers) > 0 {
			// The length will always be a list of 1, since we are matching one service name at a time
			container := containers[0]
			strategy := dc.WaitStrategyMap[k]
			dockerProvider, dpErr := NewDockerProvider()
			if dpErr != nil {
				fmt.Printf("Unable to create new Docker Provider")
				return dpErr
			}
			dockercontainer := &DockerContainer{ID: container.ID, WaitingFor: strategy, provider: dockerProvider}
			err := strategy.WaitUntilReady(context.Background(), dockercontainer)
			if err != nil {
				failedStrategies[container.Image] = strategy
				fmt.Printf("Error trace: %s\n", err)
			}
		}
	}

	if len(failedStrategies) > 0 {
		err := fmt.Errorf("List of wait strategies that were unsuccessful: (%v)", failedStrategies)
		return err
	}
	return nil
}

// Invoke invokes the docker compose
func (dc *LocalDockerCompose) Invoke() ExecError {
	if dc.waitStrategySupplied {
		return executeCompose(dc, dc.Cmd)
	}
	return executeCompose(dc, dc.Cmd)
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
// are given for a single service, the latest one will be applied
func (dc *LocalDockerCompose) WithExposedService(service string, strategy wait.Strategy) DockerCompose {
	dc.waitStrategySupplied = true
	dc.WaitStrategyMap[service] = strategy
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

		dc.Services = c.Services
	}

	return nil
}

// ExecError is super struct that holds any information about an execution error, so the client code
// can handle the result
type ExecError struct {
	Command []string
	Error   error
	Stdout  error
	Stderr  error
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
			Command: execCmd,
			Error:   err,
			Stderr:  errStderr,
			Stdout:  errStdout,
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
		Command: execCmd,
		Error:   err,
		Stderr:  errStderr,
		Stdout:  errStdout,
	}
}

func executeCompose(dc *LocalDockerCompose, args []string) ExecError {
	if which(dc.Executable) != nil {
		return ExecError{
			Command: []string{dc.Executable},
			Error:   fmt.Errorf("Local Docker Compose not found. Is %s on the PATH?", dc.Executable),
		}
	}

	environment := dc.getDockerComposeEnvironment()
	for k, v := range dc.Env {
		environment[k] = v
	}

	cmds := []string{}
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
			Command: []string{dc.Executable},
			Error:   fmt.Errorf("Local Docker compose exited abnormally whilst running %s: [%v]. %s", dc.Executable, args, err.Error()),
		}
	}

	if dc.waitStrategySupplied {
		err := dc.applyStrategyToRunningContainer()
		if err != nil {
			return ExecError{
				Error: fmt.Errorf("One or more wait strategies could not be applied to the running containers: %v", err),
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
	return w.buf.Bytes()
}

// Which checks if a binary is present in PATH
func which(binary string) error {
	_, err := exec.LookPath(binary)

	return err
}
