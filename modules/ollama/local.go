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
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

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
	// Ensure localProcess implements the required interfaces.
	_ testcontainers.Container           = (*localProcess)(nil)
	_ testcontainers.ContainerCustomizer = (*localProcess)(nil)

	// zeroTime is the zero time value.
	zeroTime time.Time
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

	// binary is the name of the Ollama binary.
	binary string
}

// runLocal returns an OllamaContainer that uses the local Ollama binary instead of using a Docker container.
func (c *localProcess) run(ctx context.Context, req testcontainers.GenericContainerRequest) (*OllamaContainer, error) {
	if err := c.validateRequest(req); err != nil {
		return nil, fmt.Errorf("validate request: %w", err)
	}

	// Apply the updated details from the request.
	c.waitFor = req.WaitingFor
	c.env = c.env[:0]
	for k, v := range req.Env {
		c.env = append(c.env, k+"="+v)
		if k == localLogVar {
			c.logName = v
		}
	}

	err := c.Start(ctx)
	var container *OllamaContainer
	if c.cmd != nil {
		container = &OllamaContainer{Container: c}
	}

	if err != nil {
		return container, fmt.Errorf("start ollama: %w", err)
	}

	return container, nil
}

// validateRequest checks that req is valid for the local Ollama binary.
func (c *localProcess) validateRequest(req testcontainers.GenericContainerRequest) error {
	var errs []error
	if req.WaitingFor == nil {
		errs = append(errs, errors.New("ContainerRequest.WaitingFor must be set"))
	}

	if !req.Started {
		errs = append(errs, errors.New("started must be true"))
	}

	if !reflect.DeepEqual(req.ExposedPorts, []string{localPort + "/tcp"}) {
		errs = append(errs, fmt.Errorf("ContainerRequest.ExposedPorts must be %s/tcp got: %s", localPort, req.ExposedPorts))
	}

	// Validate the image and extract the binary name.
	// The image must be in the format "[<path>/]<binary>[:latest]".
	if binary := req.Image; binary != "" {
		// Check if the version is "latest" or not specified.
		if idx := strings.IndexByte(binary, ':'); idx != -1 {
			if binary[idx+1:] != "latest" {
				errs = append(errs, fmt.Errorf(`ContainerRequest.Image version must be blank or "latest", got: %q`, binary[idx+1:]))
			}
			binary = binary[:idx]
		}

		// Trim the path if present.
		if idx := strings.LastIndexByte(binary, '/'); idx != -1 {
			binary = binary[idx+1:]
		}

		if _, err := exec.LookPath(binary); err != nil {
			errs = append(errs, fmt.Errorf("invalid image %q: %w", req.Image, err))
		} else {
			c.binary = binary
		}
	}

	// Reset fields we support to their zero values.
	req.Env = nil
	req.ExposedPorts = nil
	req.WaitingFor = nil
	req.Image = ""
	req.Started = false
	req.Logger = nil // We don't need the logger.

	parts := make([]string, 0, 3)
	value := reflect.ValueOf(req)
	typ := value.Type()
	fields := reflect.VisibleFields(typ)
	for _, f := range fields {
		field := value.FieldByIndex(f.Index)
		if field.Kind() == reflect.Struct {
			// Only check the leaf fields.
			continue
		}

		if !field.IsZero() {
			parts = parts[:0]
			for i := range f.Index {
				parts = append(parts, typ.FieldByIndex(f.Index[:i+1]).Name)
			}
			errs = append(errs, fmt.Errorf("unsupported field: %s = %q", strings.Join(parts, "."), field))
		}
	}

	return errors.Join(errs...)
}

// Start implements testcontainers.Container interface for the local Ollama binary.
func (c *localProcess) Start(ctx context.Context) error {
	if c.IsRunning() {
		return errors.New("already running")
	}

	cmd := exec.CommandContext(ctx, c.binary, localServeArg)
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
		return fmt.Errorf("start ollama serve: %w", errors.Join(err, c.cleanup()))
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
func (c *localProcess) extractLogDetails(pattern string, submatches [][][]byte) error {
	var err error
	for _, matches := range submatches {
		if len(matches) != 3 {
			err = fmt.Errorf("`%s` matched %d times, expected %d", pattern, len(matches), 3)
			continue
		}

		c.host, c.port, err = net.SplitHostPort(string(matches[1]))
		if err != nil {
			return wait.NewPermanentError(fmt.Errorf("split host port: %w", err))
		}

		// Set OLLAMA_HOST variable to the extracted host so Exec can use it.
		c.env = append(c.env, localHostVar+"="+string(matches[1]))
		c.version = string(matches[2])

		return nil
	}

	if err != nil {
		// Return the last error encountered.
		return err
	}

	return fmt.Errorf("address and version not found: `%s` no matches", pattern)
}

// ContainerIP implements testcontainers.Container interface for the local Ollama binary.
func (c *localProcess) ContainerIP(_ context.Context) (string, error) {
	return c.host, nil
}

// ContainerIPs returns a slice with the IP address of the local Ollama binary.
func (c *localProcess) ContainerIPs(_ context.Context) ([]string, error) {
	return []string{c.host}, nil
}

// CopyToContainer implements testcontainers.Container interface for the local Ollama binary.
// Returns [errors.ErrUnsupported].
func (c *localProcess) CopyToContainer(_ context.Context, _ []byte, _ string, _ int64) error {
	return errors.ErrUnsupported
}

// CopyDirToContainer implements testcontainers.Container interface for the local Ollama binary.
// Returns [errors.ErrUnsupported].
func (c *localProcess) CopyDirToContainer(_ context.Context, _ string, _ string, _ int64) error {
	return errors.ErrUnsupported
}

// CopyFileToContainer implements testcontainers.Container interface for the local Ollama binary.
// Returns [errors.ErrUnsupported].
func (c *localProcess) CopyFileToContainer(_ context.Context, _ string, _ string, _ int64) error {
	return errors.ErrUnsupported
}

// CopyFileFromContainer implements testcontainers.Container interface for the local Ollama binary.
// Returns [errors.ErrUnsupported].
func (c *localProcess) CopyFileFromContainer(_ context.Context, _ string) (io.ReadCloser, error) {
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
	} else if cmd[0] != c.binary {
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
func (c *localProcess) Inspect(ctx context.Context) (*container.InspectResponse, error) {
	state, err := c.State(ctx)
	if err != nil {
		return nil, fmt.Errorf("state: %w", err)
	}

	return &container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
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
			Entrypoint: []string{c.binary, localServeArg},
		},
		NetworkSettings: &container.NetworkSettings{
			Networks: map[string]*network.EndpointSettings{},
			NetworkSettingsBase: container.NetworkSettingsBase{
				Bridge: "bridge",
				Ports: nat.PortMap{
					nat.Port(localPort + "/tcp"): {
						{HostIP: c.host, HostPort: c.port},
					},
				},
			},
			DefaultNetworkSettings: container.DefaultNetworkSettings{
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
func (c *localProcess) Logs(_ context.Context) (io.ReadCloser, error) {
	file, err := os.Open(c.logFile.Name())
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return file, nil
}

// State implements testcontainers.Container interface for the local Ollama binary.
// It returns the current state of the Ollama process, simulating a container state.
func (c *localProcess) State(_ context.Context) (*container.State, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if !c.IsRunning() {
		state := &container.State{
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
	return &container.State{
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
func (c *localProcess) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	options := testcontainers.NewTerminateOptions(ctx, opts...)
	// First try to stop gracefully.
	if err := c.Stop(options.Context(), options.StopTimeout()); !c.isCleanupSafe(err) {
		return fmt.Errorf("stop: %w", err)
	}

	var errs []error
	if c.IsRunning() {
		// Still running, force kill.
		if err := c.cmd.Process.Kill(); !c.isCleanupSafe(err) {
			// Best effort so we can continue with the cleanup.
			errs = append(errs, fmt.Errorf("kill: %w", err))
		}

		// Wait for the process to exit so we can capture any error.
		c.wg.Wait()
	}

	errs = append(errs, c.cleanup(), options.Cleanup())

	return errors.Join(errs...)
}

// cleanup performs all clean up, closing and removing the log file if set.
func (c *localProcess) cleanup() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.logFile == nil {
		return c.exitErr
	}

	var errs []error
	if c.exitErr != nil {
		errs = append(errs, fmt.Errorf("exit: %w", c.exitErr))
	}

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
func (c *localProcess) Host(_ context.Context) (string, error) {
	return c.host, nil
}

// MappedPort implements testcontainers.Container interface for the local Ollama binary.
func (c *localProcess) MappedPort(_ context.Context, port nat.Port) (nat.Port, error) {
	if port.Port() != localPort || port.Proto() != "tcp" {
		return "", errdefs.NotFound(fmt.Errorf("port %q not found", port))
	}

	return nat.Port(c.port + "/tcp"), nil
}

// Networks implements testcontainers.Container interface for the local Ollama binary.
// It returns a nil slice.
func (c *localProcess) Networks(_ context.Context) ([]string, error) {
	return nil, nil
}

// NetworkAliases implements testcontainers.Container interface for the local Ollama binary.
// It returns a nil map.
func (c *localProcess) NetworkAliases(_ context.Context) (map[string][]string, error) {
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
func (c *localProcess) FollowOutput(_ testcontainers.LogConsumer) {
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

// Customize implements the [testcontainers.ContainerCustomizer] interface.
// It configures the environment variables set by [WithUseLocal] and sets up
// the wait strategy to extract the host, port and version from the log.
func (c *localProcess) Customize(req *testcontainers.GenericContainerRequest) error {
	// Replace the default host port strategy with one that waits for a log entry
	// and extracts the host, port and version from it.
	if err := wait.Walk(&req.WaitingFor, func(w wait.Strategy) error {
		if _, ok := w.(*wait.HostPortStrategy); ok {
			return wait.ErrVisitRemove
		}

		return nil
	}); err != nil {
		return fmt.Errorf("walk strategies: %w", err)
	}

	logStrategy := wait.ForLog(localLogRegex).Submatch(c.extractLogDetails)
	if req.WaitingFor == nil {
		req.WaitingFor = logStrategy
	} else {
		req.WaitingFor = wait.ForAll(req.WaitingFor, logStrategy)
	}

	// Setup the environment variables using a random port by default
	// to avoid conflicts.
	osEnv := os.Environ()
	env := make(map[string]string, len(osEnv)+len(c.env)+1)
	env[localHostVar] = "localhost:0"
	for _, kv := range append(osEnv, c.env...) {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid environment variable: %q", kv)
		}

		env[parts[0]] = parts[1]
	}

	return testcontainers.WithEnv(env)(req)
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
