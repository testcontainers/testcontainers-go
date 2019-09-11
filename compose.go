package testcontainers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
}

// LocalDockerCompose represents a Docker Compose execution using local binary
// docker-compose or docker-compose.exe, depending on the underlying platform
type LocalDockerCompose struct {
	Executable      string
	ComposeFilePath string
	Identifier      string
	Cmd             []string
	Env             map[string]string
}

// NewLocalDockerCompose returns an instance of the local Docker Compose
func NewLocalDockerCompose(filePath string, identifier string) *LocalDockerCompose {
	dc := &LocalDockerCompose{}

	dc.Executable = "docker-compose"
	if runtime.GOOS == "windows" {
		dc.Executable = "docker-compose.exe"
	}

	dc.ComposeFilePath = filePath
	dc.Identifier = identifier

	return dc
}

// Down executes docker-compose down
func (dc *LocalDockerCompose) Down() ExecError {
	if which(dc.Executable) != nil {
		panic("Local Docker Compose not found. Is " + dc.Executable + " on the PATH?")
	}

	abs, err := filepath.Abs(dc.ComposeFilePath)
	pwd, name := filepath.Split(abs)

	cmds := []string{
		"-f", name, "down",
	}

	execErr := execute(pwd, map[string]string{}, dc.Executable, cmds)
	err = execErr.Error
	if err != nil {
		args := strings.Join(dc.Cmd, " ")
		panic(
			"Local Docker compose exited abnormally whilst running " +
				dc.Executable + ": [" + args + "]. " + err.Error())
	}

	return execErr
}

// Invoke invokes the docker compose
func (dc *LocalDockerCompose) Invoke() ExecError {
	if which(dc.Executable) != nil {
		panic("Local Docker Compose not found. Is " + dc.Executable + " on the PATH?")
	}

	environment := map[string]string{}
	for k, v := range dc.Env {
		environment[k] = v
	}
	environment[envProjectName] = dc.Identifier
	environment[envComposeFile] = dc.ComposeFilePath

	abs, err := filepath.Abs(dc.ComposeFilePath)
	pwd, name := filepath.Split(abs)

	cmds := []string{
		"-f", name,
	}
	cmds = append(cmds, dc.Cmd...)

	execErr := execute(pwd, environment, dc.Executable, cmds)
	err = execErr.Error
	if err != nil {
		args := strings.Join(dc.Cmd, " ")
		panic(
			"Local Docker compose exited abnormally whilst running " +
				dc.Executable + ": [" + args + "]. " + err.Error())
	}

	return execErr
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

// ExecError is super struct that holds any information about an execution error, so the client code
// can handle the result
type ExecError struct {
	Error  error
	Stdout error
	Stderr error
}

// execute executes a program with arguments and environment variables inside a specific directory
func execute(
	dirContext string, environment map[string]string, binary string, args []string) ExecError {

	var errStdout, errStderr error

	fmt.Printf(
		"Executing %s at %s with environment: %s, and arguments: %s", binary, dirContext,
		environment, args)

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
		return ExecError{
			Error:  err,
			Stderr: errStderr,
			Stdout: errStdout,
		}
	}

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()

	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()

	err = cmd.Wait()

	return ExecError{
		Error:  err,
		Stderr: errStderr,
		Stdout: errStdout,
	}
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
