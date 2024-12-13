package ollama

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	localIP   = "127.0.0.1"
	localPort = "11434"
)

var (
	defaultStopTimeout      = time.Second * 5
	errCopyAPIsNotSupported = errors.New("copy APIs are not supported for local Ollama binary")
)

// localContext is a type holding the context for local Ollama executions.
type localContext struct {
	env      []string
	serveCmd *exec.Cmd
	logFile  *os.File
	mx       sync.Mutex
	host     string
	port     string
}

// runLocal calls the local Ollama binary instead of using a Docker container.
func runLocal(ctx context.Context, env map[string]string) (*OllamaContainer, error) {
	// Apply the environment variables to the command.
	cmdEnv := make([]string, 0, len(env)*2)
	for k, v := range env {
		cmdEnv = append(cmdEnv, k+"="+v)
	}

	localCtx := &localContext{
		env:  cmdEnv,
		host: localIP,
		port: localPort,
	}

	if envHost := os.Getenv("OLLAMA_HOST"); envHost != "" {
		host, port, err := net.SplitHostPort(envHost)
		if err != nil {
			return nil, fmt.Errorf("invalid OLLAMA_HOST: %w", err)
		}

		localCtx.host = host
		localCtx.port = port
	}

	c := &OllamaContainer{
		localCtx: localCtx,
	}

	c.localCtx.mx.Lock()

	serveCmd, logFile, err := startOllama(ctx, c.localCtx)
	if err != nil {
		c.localCtx.mx.Unlock()
		return nil, fmt.Errorf("start ollama: %w", err)
	}

	c.localCtx.serveCmd = serveCmd
	c.localCtx.logFile = logFile
	c.localCtx.mx.Unlock()

	err = c.waitForOllama(ctx)
	if err != nil {
		return nil, fmt.Errorf("wait for ollama to start: %w", err)
	}

	return c, nil
}

// logFile returns an existing log file or creates a new one if it doesn't exist.
func logFile() (*os.File, error) {
	logName := "local-ollama-" + testcontainers.SessionID() + ".log"

	if envLogName := os.Getenv("OLLAMA_LOGFILE"); envLogName != "" {
		logName = envLogName
	}

	file, err := os.Create(logName)
	if err != nil {
		return nil, fmt.Errorf("create ollama log file: %w", err)
	}

	return file, nil
}

// startOllama starts the Ollama serve command in the background, writing to the
// provided log file.
func startOllama(ctx context.Context, localCtx *localContext) (*exec.Cmd, *os.File, error) {
	serveCmd := exec.CommandContext(ctx, "ollama", "serve")
	serveCmd.Env = append(serveCmd.Env, localCtx.env...)
	serveCmd.Env = append(serveCmd.Env, os.Environ()...)

	logFile, err := logFile()
	if err != nil {
		return nil, nil, fmt.Errorf("ollama log file: %w", err)
	}

	serveCmd.Stdout = logFile
	serveCmd.Stderr = logFile

	// Run the ollama serve command in background
	err = serveCmd.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("start ollama serve: %w", err)
	}

	return serveCmd, logFile, nil
}

// waitForOllama Wait until the Ollama process is ready, checking that the log file contains
// the "Listening on 127.0.0.1:11434" message
func (c *OllamaContainer) waitForOllama(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := wait.ForLog("Listening on "+c.localCtx.host+":"+c.localCtx.port).WaitUntilReady(ctx, c)
	if err != nil {
		logs, err := c.Logs(ctx)
		if err != nil {
			return fmt.Errorf("wait for ollama to start: %w", err)
		}

		// ignore error as we already have an error and the output is already logged
		bs, _ := io.ReadAll(logs)
		return fmt.Errorf("wait for ollama to start: %w. Container logs:\n%s", err, string(bs))
	}

	return nil
}

// ContainerIP returns the IP address of the local Ollama binary.
func (c *OllamaContainer) ContainerIP(ctx context.Context) (string, error) {
	if c.localCtx == nil {
		return c.Container.ContainerIP(ctx)
	}

	return localIP, nil
}

// ContainerIPs returns a slice with the IP address of the local Ollama binary.
func (c *OllamaContainer) ContainerIPs(ctx context.Context) ([]string, error) {
	if c.localCtx == nil {
		return c.Container.ContainerIPs(ctx)
	}

	return []string{localIP}, nil
}

// CopyToContainer is a no-op for the local Ollama binary.
func (c *OllamaContainer) CopyToContainer(ctx context.Context, fileContent []byte, containerFilePath string, fileMode int64) error {
	if c.localCtx == nil {
		return c.Container.CopyToContainer(ctx, fileContent, containerFilePath, fileMode)
	}

	return errCopyAPIsNotSupported
}

// CopyDirToContainer is a no-op for the local Ollama binary.
func (c *OllamaContainer) CopyDirToContainer(ctx context.Context, hostDirPath string, containerParentPath string, fileMode int64) error {
	if c.localCtx == nil {
		return c.Container.CopyDirToContainer(ctx, hostDirPath, containerParentPath, fileMode)
	}

	return errCopyAPIsNotSupported
}

// CopyFileToContainer is a no-op for the local Ollama binary.
func (c *OllamaContainer) CopyFileToContainer(ctx context.Context, hostFilePath string, containerFilePath string, fileMode int64) error {
	if c.localCtx == nil {
		return c.Container.CopyFileToContainer(ctx, hostFilePath, containerFilePath, fileMode)
	}

	return errCopyAPIsNotSupported
}

// CopyFileFromContainer is a no-op for the local Ollama binary.
func (c *OllamaContainer) CopyFileFromContainer(ctx context.Context, filePath string) (io.ReadCloser, error) {
	if c.localCtx == nil {
		return c.Container.CopyFileFromContainer(ctx, filePath)
	}

	return nil, errCopyAPIsNotSupported
}

// GetLogProductionErrorChannel returns a nil channel.
func (c *OllamaContainer) GetLogProductionErrorChannel() <-chan error {
	if c.localCtx == nil {
		return c.Container.GetLogProductionErrorChannel()
	}

	return nil
}

// Endpoint returns the 127.0.0.1:11434 endpoint for the local Ollama binary.
func (c *OllamaContainer) Endpoint(ctx context.Context, port string) (string, error) {
	if c.localCtx == nil {
		return c.Container.Endpoint(ctx, port)
	}

	return c.localCtx.host + ":" + c.localCtx.port, nil
}

// Exec executes a command using the local Ollama binary.
func (c *OllamaContainer) Exec(ctx context.Context, cmd []string, options ...tcexec.ProcessOption) (int, io.Reader, error) {
	if c.localCtx == nil {
		return c.Container.Exec(ctx, cmd, options...)
	}

	c.localCtx.mx.Lock()
	defer c.localCtx.mx.Unlock()

	if len(cmd) == 0 {
		err := errors.New("exec: no command provided")
		return 1, strings.NewReader(err.Error()), err
	} else if cmd[0] != "ollama" {
		err := fmt.Errorf("%s: %w", cmd[0], errors.ErrUnsupported)
		return 1, strings.NewReader(err.Error()), err
	}

	args := []string{}
	if len(cmd) > 1 {
		args = cmd[1:] // prevent when there is only one command
	}

	command := prepareExec(ctx, cmd[0], args, c.localCtx.env, c.localCtx.logFile)
	err := command.Run()
	if err != nil {
		return command.ProcessState.ExitCode(), c.localCtx.logFile, fmt.Errorf("exec %v: %w", cmd, err)
	}

	return command.ProcessState.ExitCode(), c.localCtx.logFile, nil
}

func prepareExec(ctx context.Context, bin string, args []string, env []string, output io.Writer) *exec.Cmd {
	command := exec.CommandContext(ctx, bin, args...)
	command.Env = append(command.Env, env...)
	command.Env = append(command.Env, os.Environ()...)

	command.Stdout = output
	command.Stderr = output

	return command
}

// GetContainerID returns a placeholder ID for local execution
func (c *OllamaContainer) GetContainerID() string {
	if c.localCtx == nil {
		return c.Container.GetContainerID()
	}

	return "local-ollama-" + testcontainers.SessionID()
}

// Host returns the 127.0.0.1 address for the local Ollama binary.
func (c *OllamaContainer) Host(ctx context.Context) (string, error) {
	if c.localCtx == nil {
		return c.Container.Host(ctx)
	}

	return localIP, nil
}

// Inspect returns a ContainerJSON with the state of the local Ollama binary.
// The version is read from the local Ollama binary (ollama -v), and the port
// mapping is set to 11434.
func (c *OllamaContainer) Inspect(ctx context.Context) (*types.ContainerJSON, error) {
	if c.localCtx == nil {
		return c.Container.Inspect(ctx)
	}

	state, err := c.State(ctx)
	if err != nil {
		return nil, fmt.Errorf("get ollama state: %w", err)
	}

	// read the version from the ollama binary
	var buf bytes.Buffer
	command := prepareExec(ctx, "ollama", []string{"-v"}, c.localCtx.env, &buf)
	err = command.Run()
	if err != nil {
		return nil, fmt.Errorf("read ollama -v output: %w", err)
	}

	bs, err := io.ReadAll(&buf)
	if err != nil {
		return nil, fmt.Errorf("read ollama -v output: %w", err)
	}

	return &types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:    c.GetContainerID(),
			Name:  "local-ollama-" + testcontainers.SessionID(),
			State: state,
		},
		Config: &container.Config{
			Image: string(bs),
			ExposedPorts: nat.PortSet{
				nat.Port(c.localCtx.port + "/tcp"): struct{}{},
			},
			Hostname:   "localhost",
			Entrypoint: []string{"ollama", "serve"},
		},
		NetworkSettings: &types.NetworkSettings{
			Networks: map[string]*network.EndpointSettings{},
			NetworkSettingsBase: types.NetworkSettingsBase{
				Bridge: "bridge",
				Ports: nat.PortMap{
					nat.Port(c.localCtx.port + "/tcp"): {
						{HostIP: c.localCtx.host, HostPort: c.localCtx.port},
					},
				},
			},
			DefaultNetworkSettings: types.DefaultNetworkSettings{
				IPAddress: c.localCtx.host,
			},
		},
	}, nil
}

// IsRunning returns true if the local Ollama process is running.
func (c *OllamaContainer) IsRunning() bool {
	if c.localCtx == nil {
		return c.Container.IsRunning()
	}

	c.localCtx.mx.Lock()
	defer c.localCtx.mx.Unlock()

	return c.localCtx.serveCmd != nil
}

// Logs returns the logs from the local Ollama binary.
func (c *OllamaContainer) Logs(ctx context.Context) (io.ReadCloser, error) {
	if c.localCtx == nil {
		return c.Container.Logs(ctx)
	}

	c.localCtx.mx.Lock()
	defer c.localCtx.mx.Unlock()

	// stream the log file
	return os.Open(c.localCtx.logFile.Name())
}

// MappedPort returns the configured port for local Ollama binary.
func (c *OllamaContainer) MappedPort(ctx context.Context, port nat.Port) (nat.Port, error) {
	if c.localCtx == nil {
		return c.Container.MappedPort(ctx, port)
	}

	// Ollama typically uses port 11434 by default
	return nat.Port(c.localCtx.port + "/tcp"), nil
}

// Networks returns the networks for local Ollama binary, which is a nil slice.
func (c *OllamaContainer) Networks(ctx context.Context) ([]string, error) {
	if c.localCtx == nil {
		return c.Container.Networks(ctx)
	}

	return nil, nil
}

// NetworkAliases returns the network aliases for local Ollama binary, which is a nil map.
func (c *OllamaContainer) NetworkAliases(ctx context.Context) (map[string][]string, error) {
	if c.localCtx == nil {
		return c.Container.NetworkAliases(ctx)
	}

	return nil, nil
}

// SessionID returns the session ID for local Ollama binary, which is the session ID
// of the test execution.
func (c *OllamaContainer) SessionID() string {
	if c.localCtx == nil {
		return c.Container.SessionID()
	}

	return testcontainers.SessionID()
}

// Start starts the local Ollama process, not failing if it's already running.
func (c *OllamaContainer) Start(ctx context.Context) error {
	if c.localCtx == nil {
		return c.Container.Start(ctx)
	}

	c.localCtx.mx.Lock()

	if c.localCtx.serveCmd != nil {
		c.localCtx.mx.Unlock()
		return nil
	}

	serveCmd, logFile, err := startOllama(ctx, c.localCtx)
	if err != nil {
		c.localCtx.mx.Unlock()
		return fmt.Errorf("start ollama: %w", err)
	}
	c.localCtx.serveCmd = serveCmd
	c.localCtx.logFile = logFile
	c.localCtx.mx.Unlock() // unlock before waiting for the process to be ready

	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = c.waitForOllama(waitCtx)
	if err != nil {
		return fmt.Errorf("wait for ollama to start: %w", err)
	}

	return nil
}

// State returns the current state of the Ollama process, simulating a container state
// for local execution.
func (c *OllamaContainer) State(ctx context.Context) (*types.ContainerState, error) {
	if c.localCtx == nil {
		return c.Container.State(ctx)
	}

	c.localCtx.mx.Lock()
	defer c.localCtx.mx.Unlock()

	if c.localCtx.serveCmd == nil {
		return &types.ContainerState{Status: "stopped"}, nil
	}

	// Check if process is still running. Signal(0) is a special case in Unix-like systems.
	// When you send signal 0 to a process:
	// - It performs all the normal error checking (permissions, process existence, etc.)
	// - But it doesn't actually send any signal to the process
	if err := c.localCtx.serveCmd.Process.Signal(syscall.Signal(0)); err != nil {
		return &types.ContainerState{Status: "stopped"}, nil
	}

	// Setting the Running field because it's required by the wait strategy
	// to check if the given log message is present.
	return &types.ContainerState{Status: "running", Running: true}, nil
}

// Stop gracefully stops the local Ollama process
func (c *OllamaContainer) Stop(ctx context.Context, d *time.Duration) error {
	if c.localCtx == nil {
		return c.Container.Stop(ctx, d)
	}

	c.localCtx.mx.Lock()
	defer c.localCtx.mx.Unlock()

	if c.localCtx.serveCmd == nil {
		return nil
	}

	if err := c.localCtx.serveCmd.Process.Signal(syscall.Signal(syscall.SIGTERM)); err != nil {
		return fmt.Errorf("signal ollama: %w", err)
	}

	c.localCtx.serveCmd = nil

	return nil
}

// Terminate stops the local Ollama process, removing the log file.
func (c *OllamaContainer) Terminate(ctx context.Context) error {
	if c.localCtx == nil {
		return c.Container.Terminate(ctx)
	}

	// First try to stop gracefully
	err := c.Stop(ctx, &defaultStopTimeout)
	if err != nil {
		return fmt.Errorf("stop ollama: %w", err)
	}

	c.localCtx.mx.Lock()
	defer c.localCtx.mx.Unlock()

	if c.localCtx.logFile == nil {
		return nil
	}

	// remove the log file if it exists
	if _, err = os.Stat(c.localCtx.logFile.Name()); err == nil {
		err = c.localCtx.logFile.Close()
		if err != nil {
			return err
		}

		err = os.Remove(c.localCtx.logFile.Name())
		if err != nil {
			return err
		}
	}

	return nil
}
