package ollama

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	localPort       = "11434"
	localBinary     = "ollama"
	localServeArg   = "serve"
	localLogRegex   = `Listening on (.*:\d+) \(version\s(.*)\)`
	localNamePrefix = "local-ollama"
	localHostVar    = "OLLAMA_HOST"
	localLogVar     = "OLLAMA_LOGFILE"
)

var (
	// Ensure localContext implements the testcontainers.Container interface.
	_ testcontainers.Container = &localProcess{}

	// defaultStopTimeout is the default timeout for stopping the local Ollama process.
	defaultStopTimeout = time.Second * 5

	// zeroTime is the zero time value.
	zeroTime time.Time

	// reLogDetails is the regular expression to extract the listening address and version from the log.
	reLogDetails = regexp.MustCompile(localLogRegex)
)

// localProcess emulates the Ollama container using a local process to improve performance.
type localProcess struct {
	sessionID string

	// env is the combined environment variables passed to the Ollama binary.
	env []string

	// cmd is the command that runs the Ollama binary, not valid externally if nil.
	cmd *exec.Cmd

	// logName and logFile are the file where the Ollama logs are written.
	logName string
	logFile *os.File

	// host, port and version are extracted from log on startup.
	host    string
	port    string
	version string

	// waitFor is the strategy to wait for the process to be ready.
	waitFor wait.Strategy

	// done is closed when the process is finished.
	done chan struct{}

	// wg is used to wait for the process to finish.
	wg sync.WaitGroup

	// startedAt is the time when the process started.
	startedAt time.Time

	// mtx is used to synchronize access to the process state fields below.
	mtx sync.Mutex

	// finishedAt is the time when the process finished.
	finishedAt time.Time

	// exitErr is the error returned by the process.
	exitErr error
}

// runLocal returns an OllamaContainer that uses the local Ollama binary instead of using a Docker container.
func runLocal(ctx context.Context, req testcontainers.GenericContainerRequest) (*OllamaContainer, error) {
	// TODO: validate the request and return an error if it
	// contains any unsupported elements.

	sessionID := testcontainers.SessionID()
	local := &localProcess{
		sessionID: sessionID,
		env:       make([]string, 0, len(req.Env)),
		waitFor:   req.WaitingFor,
		logName:   localNamePrefix + "-" + sessionID + ".log",
	}

	// Apply the environment variables to the command and
	// override the log file if specified.
	for k, v := range req.Env {
		local.env = append(local.env, k+"="+v)
		if k == localLogVar {
			local.logName = v
		}
	}

	err := local.Start(ctx)
	var c *OllamaContainer
	if local.cmd != nil {
		c = &OllamaContainer{Container: local}
	}

	if err != nil {
		return nil, fmt.Errorf("start ollama: %w", err)
	}

	return c, nil
}

// Start implements testcontainers.Container interface for the local Ollama binary.
func (c *localProcess) Start(ctx context.Context) error {
	if c.IsRunning() {
		return errors.New("already running")
	}

	cmd := exec.CommandContext(ctx, localBinary, localServeArg)
	cmd.Env = c.env

	var err error
	c.logFile, err = os.Create(c.logName)
	if err != nil {
		return fmt.Errorf("create ollama log file: %w", err)
	}

	// Multiplex stdout and stderr to the log file matching the Docker API.
	cmd.Stdout = stdcopy.NewStdWriter(c.logFile, stdcopy.Stdout)
	cmd.Stderr = stdcopy.NewStdWriter(c.logFile, stdcopy.Stderr)

	// Run the ollama serve command in background.
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("start ollama serve: %w", errors.Join(err, c.cleanupLog()))
	}

	// Past this point, the process was started successfully.
	c.cmd = cmd
	c.startedAt = time.Now()

	// Reset the details to allow multiple start / stop cycles.
	c.done = make(chan struct{})
	c.mtx.Lock()
	c.finishedAt = zeroTime
	c.exitErr = nil
	c.mtx.Unlock()

	// Wait for the process to finish in a goroutine.
	c.wg.Add(1)
	go func() {
		defer func() {
			c.wg.Done()
			close(c.done)
		}()

		err := c.cmd.Wait()
		c.mtx.Lock()
		defer c.mtx.Unlock()
		if err != nil {
			c.exitErr = fmt.Errorf("process wait: %w", err)
		}
		c.finishedAt = time.Now()
	}()

	if err = c.waitStrategy(ctx); err != nil {
		return fmt.Errorf("wait strategy: %w", err)
	}

	if err := c.extractLogDetails(ctx); err != nil {
		return fmt.Errorf("extract log details: %w", err)
	}

	return nil
}

// waitStrategy waits until the Ollama process is ready.
func (c *localProcess) waitStrategy(ctx context.Context) error {
	if err := c.waitFor.WaitUntilReady(ctx, c); err != nil {
		logs, lerr := c.Logs(ctx)
		if lerr != nil {
			return errors.Join(err, lerr)
		}
		defer logs.Close()

		var stderr, stdout bytes.Buffer
		_, cerr := stdcopy.StdCopy(&stdout, &stderr, logs)

		return fmt.Errorf(
			"%w (stdout: %s, stderr: %s)",
			errors.Join(err, cerr),
			strings.TrimSpace(stdout.String()),
			strings.TrimSpace(stderr.String()),
		)
	}

	return nil
}

// extractLogDetails extracts the listening address and version from the log.
func (c *localProcess) extractLogDetails(ctx context.Context) error {
	rc, err := c.Logs(ctx)
	if err != nil {
		return fmt.Errorf("logs: %w", err)
	}
	defer rc.Close()

	bs, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("read logs: %w", err)
	}

	matches := reLogDetails.FindSubmatch(bs)
	if len(matches) != 3 {
		return errors.New("address and version not found")
	}

	c.host, c.port, err = net.SplitHostPort(string(matches[1]))
	if err != nil {
		return fmt.Errorf("split host port: %w", err)
	}

	// Set OLLAMA_HOST variable to the extracted host so Exec can use it.
	c.env = append(c.env, localHostVar+"="+string(matches[1]))
	c.version = string(matches[2])

	return nil
}

// ContainerIP implements testcontainers.Container interface for the local Ollama binary.
func (c *localProcess) ContainerIP(ctx context.Context) (string, error) {
	return c.host, nil
}

// ContainerIPs returns a slice with the IP address of the local Ollama binary.
func (c *localProcess) ContainerIPs(ctx context.Context) ([]string, error) {
	return []string{c.host}, nil
}

// CopyToContainer implements testcontainers.Container interface for the local Ollama binary.
// Returns [errors.ErrUnsupported].
func (c *localProcess) CopyToContainer(ctx context.Context, fileContent []byte, containerFilePath string, fileMode int64) error {
	return errors.ErrUnsupported
}

// CopyDirToContainer implements testcontainers.Container interface for the local Ollama binary.
// Returns [errors.ErrUnsupported].
func (c *localProcess) CopyDirToContainer(ctx context.Context, hostDirPath string, containerParentPath string, fileMode int64) error {
	return errors.ErrUnsupported
}

// CopyFileToContainer implements testcontainers.Container interface for the local Ollama binary.
// Returns [errors.ErrUnsupported].
func (c *localProcess) CopyFileToContainer(ctx context.Context, hostFilePath string, containerFilePath string, fileMode int64) error {
	return errors.ErrUnsupported
}

// CopyFileFromContainer implements testcontainers.Container interface for the local Ollama binary.
// Returns [errors.ErrUnsupported].
func (c *localProcess) CopyFileFromContainer(ctx context.Context, filePath string) (io.ReadCloser, error) {
	return nil, errors.ErrUnsupported
}

// GetLogProductionErrorChannel implements testcontainers.Container interface for the local Ollama binary.
// It returns a nil channel because the local Ollama binary doesn't have a production error channel.
func (c *localProcess) GetLogProductionErrorChannel() <-chan error {
	return nil
}

// Exec implements testcontainers.Container interface for the local Ollama binary.
// It executes a command using the local Ollama binary and returns the exit status
// of the executed command, an [io.Reader] containing the combined stdout and stderr,
// and any encountered error.
//
// Reading directly from the [io.Reader] may result in unexpected bytes due to custom
// stream multiplexing headers. Use [tcexec.Multiplexed] option to read the combined output
// without the multiplexing headers.
// Alternatively, to separate the stdout and stderr from [io.Reader] and interpret these
// headers properly, [stdcopy.StdCopy] from the Docker API should be used.
func (c *localProcess) Exec(ctx context.Context, cmd []string, options ...tcexec.ProcessOption) (int, io.Reader, error) {
	if len(cmd) == 0 {
		return 1, nil, errors.New("no command provided")
	} else if cmd[0] != localBinary {
		return 1, nil, fmt.Errorf("command %q: %w", cmd[0], errors.ErrUnsupported)
	}

	command := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	command.Env = c.env

	// Multiplex stdout and stderr to the buffer so they can be read separately later.
	var buf bytes.Buffer
	command.Stdout = stdcopy.NewStdWriter(&buf, stdcopy.Stdout)
	command.Stderr = stdcopy.NewStdWriter(&buf, stdcopy.Stderr)

	// Use process options to customize the command execution
	// emulating the Docker API behaviour.
	processOptions := tcexec.NewProcessOptions(cmd)
	processOptions.Reader = &buf
	for _, o := range options {
		o.Apply(processOptions)
	}

	if err := c.validateExecOptions(processOptions.ExecConfig); err != nil {
		return 1, nil, fmt.Errorf("validate exec option: %w", err)
	}

	if !processOptions.ExecConfig.AttachStderr {
		command.Stderr = io.Discard
	}
	if !processOptions.ExecConfig.AttachStdout {
		command.Stdout = io.Discard
	}
	if processOptions.ExecConfig.AttachStdin {
		command.Stdin = os.Stdin
	}

	command.Dir = processOptions.ExecConfig.WorkingDir
	command.Env = append(command.Env, processOptions.ExecConfig.Env...)

	if err := command.Run(); err != nil {
		return command.ProcessState.ExitCode(), processOptions.Reader, fmt.Errorf("exec %v: %w", cmd, err)
	}

	return command.ProcessState.ExitCode(), processOptions.Reader, nil
}

// validateExecOptions checks if the given exec options are supported by the local Ollama binary.
func (c *localProcess) validateExecOptions(options container.ExecOptions) error {
	var errs []error
	if options.User != "" {
		errs = append(errs, fmt.Errorf("user: %w", errors.ErrUnsupported))
	}
	if options.Privileged {
		errs = append(errs, fmt.Errorf("privileged: %w", errors.ErrUnsupported))
	}
	if options.Tty {
		errs = append(errs, fmt.Errorf("tty: %w", errors.ErrUnsupported))
	}
	if options.Detach {
		errs = append(errs, fmt.Errorf("detach: %w", errors.ErrUnsupported))
	}
	if options.DetachKeys != "" {
		errs = append(errs, fmt.Errorf("detach keys: %w", errors.ErrUnsupported))
	}

	return errors.Join(errs...)
}

// Inspect implements testcontainers.Container interface for the local Ollama binary.
// It returns a ContainerJSON with the state of the local Ollama binary.
func (c *localProcess) Inspect(ctx context.Context) (*types.ContainerJSON, error) {
	state, err := c.State(ctx)
	if err != nil {
		return nil, fmt.Errorf("state: %w", err)
	}

	return &types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:    c.GetContainerID(),
			Name:  localNamePrefix + "-" + c.sessionID,
			State: state,
		},
		Config: &container.Config{
			Image: localNamePrefix + ":" + c.version,
			ExposedPorts: nat.PortSet{
				nat.Port(localPort + "/tcp"): struct{}{},
			},
			Hostname:   c.host,
			Entrypoint: []string{localBinary, localServeArg},
		},
		NetworkSettings: &types.NetworkSettings{
			Networks: map[string]*network.EndpointSettings{},
			NetworkSettingsBase: types.NetworkSettingsBase{
				Bridge: "bridge",
				Ports: nat.PortMap{
					nat.Port(localPort + "/tcp"): {
						{HostIP: c.host, HostPort: c.port},
					},
				},
			},
			DefaultNetworkSettings: types.DefaultNetworkSettings{
				IPAddress: c.host,
			},
		},
	}, nil
}

// IsRunning implements testcontainers.Container interface for the local Ollama binary.
// It returns true if the local Ollama process is running, false otherwise.
func (c *localProcess) IsRunning() bool {
	if c.startedAt.IsZero() {
		// The process hasn't started yet.
		return false
	}

	select {
	case <-c.done:
		// The process exited.
		return false
	default:
		// The process is still running.
		return true
	}
}

// Logs implements testcontainers.Container interface for the local Ollama binary.
// It returns the logs from the local Ollama binary.
func (c *localProcess) Logs(ctx context.Context) (io.ReadCloser, error) {
	file, err := os.Open(c.logFile.Name())
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return file, nil
}

// State implements testcontainers.Container interface for the local Ollama binary.
// It returns the current state of the Ollama process, simulating a container state.
func (c *localProcess) State(ctx context.Context) (*types.ContainerState, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if !c.IsRunning() {
		state := &types.ContainerState{
			Status:     "exited",
			ExitCode:   c.cmd.ProcessState.ExitCode(),
			StartedAt:  c.startedAt.Format(time.RFC3339Nano),
			FinishedAt: c.finishedAt.Format(time.RFC3339Nano),
		}
		if c.exitErr != nil {
			state.Error = c.exitErr.Error()
		}

		return state, nil
	}

	// Setting the Running field because it's required by the wait strategy
	// to check if the given log message is present.
	return &types.ContainerState{
		Status:     "running",
		Running:    true,
		Pid:        c.cmd.Process.Pid,
		StartedAt:  c.startedAt.Format(time.RFC3339Nano),
		FinishedAt: c.finishedAt.Format(time.RFC3339Nano),
	}, nil
}

// Stop implements testcontainers.Container interface for the local Ollama binary.
// It gracefully stops the local Ollama process.
func (c *localProcess) Stop(ctx context.Context, d *time.Duration) error {
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal ollama: %w", err)
	}

	if d != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *d)
		defer cancel()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		// The process exited.
		c.mtx.Lock()
		defer c.mtx.Unlock()

		return c.exitErr
	}
}

// Terminate implements testcontainers.Container interface for the local Ollama binary.
// It stops the local Ollama process, removing the log file.
func (c *localProcess) Terminate(ctx context.Context) error {
	// First try to stop gracefully.
	if err := c.Stop(ctx, &defaultStopTimeout); !c.isCleanupSafe(err) {
		return fmt.Errorf("stop: %w", err)
	}

	if c.IsRunning() {
		// Still running, force kill.
		if err := c.cmd.Process.Kill(); !c.isCleanupSafe(err) {
			return fmt.Errorf("kill: %w", err)
		}

		// Wait for the process to exit so capture any error.
		c.wg.Wait()
	}

	c.mtx.Lock()
	exitErr := c.exitErr
	c.mtx.Unlock()

	return errors.Join(exitErr, c.cleanupLog())
}

// cleanupLog closes the log file and removes it.
func (c *localProcess) cleanupLog() error {
	if c.logFile == nil {
		return nil
	}

	var errs []error
	if err := c.logFile.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close log: %w", err))
	}

	if err := os.Remove(c.logFile.Name()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		errs = append(errs, fmt.Errorf("remove log: %w", err))
	}

	c.logFile = nil // Prevent double cleanup.

	return errors.Join(errs...)
}

// Endpoint implements testcontainers.Container interface for the local Ollama binary.
// It returns proto://host:port string for the Ollama port.
// It returns just host:port if proto is blank.
func (c *localProcess) Endpoint(ctx context.Context, proto string) (string, error) {
	return c.PortEndpoint(ctx, localPort, proto)
}

// GetContainerID implements testcontainers.Container interface for the local Ollama binary.
func (c *localProcess) GetContainerID() string {
	return localNamePrefix + "-" + c.sessionID
}

// Host implements testcontainers.Container interface for the local Ollama binary.
func (c *localProcess) Host(ctx context.Context) (string, error) {
	return c.host, nil
}

// MappedPort implements testcontainers.Container interface for the local Ollama binary.
func (c *localProcess) MappedPort(ctx context.Context, port nat.Port) (nat.Port, error) {
	if port.Port() != localPort || port.Proto() != "tcp" {
		return "", errdefs.NotFound(fmt.Errorf("port %q not found", port))
	}

	return nat.Port(c.port + "/tcp"), nil
}

// Networks implements testcontainers.Container interface for the local Ollama binary.
// It returns a nil slice.
func (c *localProcess) Networks(ctx context.Context) ([]string, error) {
	return nil, nil
}

// NetworkAliases implements testcontainers.Container interface for the local Ollama binary.
// It returns a nil map.
func (c *localProcess) NetworkAliases(ctx context.Context) (map[string][]string, error) {
	return nil, nil
}

// PortEndpoint implements testcontainers.Container interface for the local Ollama binary.
// It returns proto://host:port string for the given exposed port.
// It returns just host:port if proto is blank.
func (c *localProcess) PortEndpoint(ctx context.Context, port nat.Port, proto string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	outerPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	if proto != "" {
		proto += "://"
	}

	return fmt.Sprintf("%s%s:%s", proto, host, outerPort.Port()), nil
}

// SessionID implements testcontainers.Container interface for the local Ollama binary.
func (c *localProcess) SessionID() string {
	return c.sessionID
}

// Deprecated: it will be removed in the next major release.
// FollowOutput is not implemented for the local Ollama binary.
// It panics if called.
func (c *localProcess) FollowOutput(consumer testcontainers.LogConsumer) {
	panic("not implemented")
}

// Deprecated: use c.Inspect(ctx).NetworkSettings.Ports instead.
// Ports gets the exposed ports for the container.
func (c *localProcess) Ports(ctx context.Context) (nat.PortMap, error) {
	inspect, err := c.Inspect(ctx)
	if err != nil {
		return nil, err
	}

	return inspect.NetworkSettings.Ports, nil
}

// Deprecated: it will be removed in the next major release.
// StartLogProducer implements testcontainers.Container interface for the local Ollama binary.
// It returns an error because the local Ollama binary doesn't have a log producer.
func (c *localProcess) StartLogProducer(context.Context, ...testcontainers.LogProductionOption) error {
	return errors.ErrUnsupported
}

// Deprecated: it will be removed in the next major release.
// StopLogProducer implements testcontainers.Container interface for the local Ollama binary.
// It returns an error because the local Ollama binary doesn't have a log producer.
func (c *localProcess) StopLogProducer() error {
	return errors.ErrUnsupported
}

// Deprecated: Use c.Inspect(ctx).Name instead.
// Name returns the name for the local Ollama binary.
func (c *localProcess) Name(context.Context) (string, error) {
	return localNamePrefix + "-" + c.sessionID, nil
}

// isCleanupSafe reports whether all errors in err's tree are one of the
// following, so can safely be ignored:
//   - nil
//   - os: process already finished
//   - context deadline exceeded
func (c *localProcess) isCleanupSafe(err error) bool {
	switch {
	case err == nil,
		errors.Is(err, os.ErrProcessDone),
		errors.Is(err, context.DeadlineExceeded):
		return true
	default:
		return false
	}
}
