package testcontainers

import (
	"archive/tar"
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/go-connections/nat"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/internal/file"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Implement interfaces
var (
	_ Container = (Container)(nil)
	_ Container = (CreatedContainer)(nil)
	_ Container = (ReadyContainer)(nil)
	_ Container = (StartedContainer)(nil)
)

// DockerContainer represents a container started using Docker
type DockerContainer struct {
	ID                string
	raw               *types.ContainerJSON
	isRunning         bool
	WaitingFor        wait.Strategy
	logger            log.Logging
	Image             string
	imageWasBuilt     bool
	lifecycleHooks    []LifecycleHooks
	sessionID         string
	terminationSignal chan bool
	keepBuiltImage    bool
	healthStatus      string // container health status, will default to healthStatusNone if no healthcheck is present

	logProductionStop chan struct{}

	logProductionTimeout *time.Duration
	logProductionError   chan error
}

// containerFromDockerResponse builds a Docker container struct from the response of the Docker API
func containerFromDockerResponse(ctx context.Context, response types.Container) (*DockerContainer, error) {
	ctr := DockerContainer{}

	ctr.ID = response.ID
	ctr.WaitingFor = nil
	ctr.Image = response.Image
	ctr.imageWasBuilt = false

	// TODO define a logger for the library
	// ctr.logger = provider.Logger
	ctr.lifecycleHooks = []LifecycleHooks{
		DefaultLoggingHook(ctr.logger),
	}

	ctr.logger = log.StandardLogger() // assign the standard logger to the container
	ctr.sessionID = core.SessionID()
	ctr.isRunning = response.State == "running"

	// the termination signal should be obtained from the reaper
	ctr.terminationSignal = nil

	// populate the raw representation of the container
	_, err := ctr.inspectRawContainer(ctx)
	if err != nil {
		return nil, err
	}

	// the health status of the container, if any
	if health := ctr.raw.State.Health; health != nil {
		ctr.healthStatus = health.Status
	}

	return &ctr, nil
}

// ContainerIP gets the IP address of the primary network within the container.
func (c *DockerContainer) ContainerIP(ctx context.Context) (string, error) {
	inspect, err := c.Inspect(ctx)
	if err != nil {
		return "", err
	}

	ip := inspect.NetworkSettings.IPAddress
	if ip == "" {
		// use IP from "Networks" if only single network defined
		networks := inspect.NetworkSettings.Networks
		if len(networks) == 1 {
			for _, v := range networks {
				ip = v.IPAddress
			}
		}
	}

	return ip, nil
}

// ContainerIPs gets the IP addresses of all the networks within the container.
func (c *DockerContainer) ContainerIPs(ctx context.Context) ([]string, error) {
	ips := make([]string, 0)

	inspect, err := c.Inspect(ctx)
	if err != nil {
		return nil, err
	}

	networks := inspect.NetworkSettings.Networks
	for _, nw := range networks {
		ips = append(ips, nw.IPAddress)
	}

	return ips, nil
}

// CopyDirToContainer copies the contents of a directory to a parent path in the container. This parent path must exist in the container first
// as we cannot create it
func (c *DockerContainer) CopyDirToContainer(ctx context.Context, hostDirPath string, containerParentPath string, fileMode int64) error {
	dir, err := file.IsDir(hostDirPath)
	if err != nil {
		return err
	}

	if !dir {
		// it's not a dir: let the consumer to handle an error
		return fmt.Errorf("path %s is not a directory", hostDirPath)
	}

	buff, err := file.TarDir(hostDirPath, fileMode)
	if err != nil {
		return err
	}

	// create the directory under its parent
	parent := filepath.Dir(containerParentPath)

	cli, err := core.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()

	err = cli.CopyToContainer(ctx, c.ID, parent, buff, types.CopyToContainerOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *DockerContainer) CopyFileFromContainer(ctx context.Context, filePath string) (io.ReadCloser, error) {
	cli, err := core.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	r, _, err := cli.CopyFromContainer(ctx, c.ID, filePath)
	if err != nil {
		return nil, err
	}

	tarReader := tar.NewReader(r)

	// if we got here we have exactly one file in the TAR-stream
	// so we advance the index by one so the next call to Read will start reading it
	_, err = tarReader.Next()
	if err != nil {
		return nil, err
	}

	ret := &FileFromContainer{
		Underlying: &r,
		Tarreader:  tarReader,
	}

	return ret, nil
}

func (c *DockerContainer) CopyFileToContainer(ctx context.Context, hostFilePath string, containerFilePath string, fileMode int64) error {
	dir, err := file.IsDir(hostFilePath)
	if err != nil {
		return err
	}

	if dir {
		return c.CopyDirToContainer(ctx, hostFilePath, containerFilePath, fileMode)
	}

	f, err := os.Open(hostFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	// In Go 1.22 os.File is always an io.WriterTo. However, testcontainers
	// currently allows Go 1.21, so we need to trick the compiler a little.
	var file fs.File = f
	return c.copyToContainer(ctx, func(tw io.Writer) error {
		// Attempt optimized writeTo, implemented in linux
		if wt, ok := file.(io.WriterTo); ok {
			_, err := wt.WriteTo(tw)
			return err
		}
		_, err := io.Copy(tw, f)
		return err
	}, info.Size(), containerFilePath, fileMode)
}

// CopyToContainer copies fileContent data to a file in container
func (c *DockerContainer) CopyToContainer(ctx context.Context, fileContent []byte, containerFilePath string, fileMode int64) error {
	return c.copyToContainer(ctx, func(tw io.Writer) error {
		_, err := tw.Write(fileContent)
		return err
	}, int64(len(fileContent)), containerFilePath, fileMode)
}

func (c *DockerContainer) copyToContainer(ctx context.Context, fileContent func(tw io.Writer) error, fileContentSize int64, containerFilePath string, fileMode int64) error {
	buffer, err := file.TarFile(containerFilePath, fileContent, fileContentSize, fileMode)
	if err != nil {
		return err
	}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()

	err = cli.CopyToContainer(ctx, c.ID, "/", buffer, types.CopyToContainerOptions{})
	if err != nil {
		return err
	}

	return nil
}

// Endpoint gets proto://host:port string for the first exposed port
// Will returns just host:port if proto is ""
func (c *DockerContainer) Endpoint(ctx context.Context, proto string) (string, error) {
	inspect, err := c.Inspect(ctx)
	if err != nil {
		return "", err
	}

	ports := inspect.NetworkSettings.Ports

	// get first port
	var firstPort nat.Port
	for p := range ports {
		firstPort = p
		break
	}

	return c.PortEndpoint(ctx, firstPort, proto)
}

// Exec executes a command in the current container.
// It returns the exit status of the executed command, an [io.Reader] containing the combined
// stdout and stderr, and any encountered error. Note that reading directly from the [io.Reader]
// may result in unexpected bytes due to custom stream multiplexing headers.
// Use [tcexec.Multiplexed] option to read the combined output without the multiplexing headers.
// Alternatively, to separate the stdout and stderr from [io.Reader] and interpret these headers properly,
// [github.com/docker/docker/pkg/stdcopy.StdCopy] from the Docker API should be used.
func (c *DockerContainer) Exec(ctx context.Context, cmd []string, options ...tcexec.ProcessOption) (int, io.Reader, error) {
	cli, err := core.NewClient(ctx)
	if err != nil {
		return 0, nil, err
	}
	defer cli.Close()

	processOptions := tcexec.NewProcessOptions(cmd)

	// processing all the options in a first loop because for the multiplexed option
	// we first need to have a containerExecCreateResponse
	for _, o := range options {
		o.Apply(processOptions)
	}

	response, err := cli.ContainerExecCreate(ctx, c.ID, processOptions.ExecConfig)
	if err != nil {
		return 0, nil, err
	}

	hijack, err := cli.ContainerExecAttach(ctx, response.ID, types.ExecStartCheck{})
	if err != nil {
		return 0, nil, err
	}

	processOptions.Reader = hijack.Reader

	// second loop to process the multiplexed option, as now we have a reader
	// from the created exec response.
	for _, o := range options {
		o.Apply(processOptions)
	}

	var exitCode int
	for {
		execResp, err := cli.ContainerExecInspect(ctx, response.ID)
		if err != nil {
			return 0, nil, err
		}

		if !execResp.Running {
			exitCode = execResp.ExitCode
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	return exitCode, processOptions.Reader, nil
}

func (c *DockerContainer) GetImage() string {
	jsonRaw, err := c.inspectRawContainer(context.Background())
	if err != nil {
		return ""
	}

	return jsonRaw.Config.Image
}

// Host gets host (ip or name) of the docker daemon where the container port is exposed
// Warning: this is based on your Docker host setting. Will fail if using an SSH tunnel
// You can use the "TC_HOST" env variable to set this yourself
func (c *DockerContainer) Host(ctx context.Context) (string, error) {
	host, err := DaemonHost(ctx)
	if err != nil {
		return "", err
	}
	return host, nil
}

// Inspect gets the raw container info, caching the result for subsequent calls
func (c *DockerContainer) Inspect(ctx context.Context) (*types.ContainerJSON, error) {
	if c.raw != nil {
		return c.raw, nil
	}

	jsonRaw, err := c.inspectRawContainer(ctx)
	if err != nil {
		return nil, err
	}

	return jsonRaw, nil
}

// update container raw info
func (c *DockerContainer) inspectRawContainer(ctx context.Context) (*types.ContainerJSON, error) {
	cli, err := core.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	inspect, err := cli.ContainerInspect(ctx, c.ID)
	if err != nil {
		return nil, err
	}

	c.raw = &inspect
	return c.raw, nil
}

func (c *DockerContainer) IsRunning() bool {
	return c.isRunning
}

// Logs will fetch both STDOUT and STDERR from the current container. Returns a
// ReadCloser and leaves it up to the caller to extract what it wants.
func (c *DockerContainer) Logs(ctx context.Context) (io.ReadCloser, error) {
	const streamHeaderSize = 8

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	rc, err := cli.ContainerLogs(ctx, c.ID, options)
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	r := bufio.NewReader(rc)

	go func() {
		lineStarted := true
		for err == nil {
			line, isPrefix, err := r.ReadLine()

			if lineStarted && len(line) >= streamHeaderSize {
				line = line[streamHeaderSize:] // trim stream header
				lineStarted = false
			}
			if !isPrefix {
				lineStarted = true
			}

			_, errW := pw.Write(line)
			if errW != nil {
				return
			}

			if !isPrefix {
				_, errW := pw.Write([]byte("\n"))
				if errW != nil {
					return
				}
			}

			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}
		}
	}()

	return pr, nil
}

// MappedPort gets externally mapped port for a container port
func (c *DockerContainer) MappedPort(ctx context.Context, port nat.Port) (nat.Port, error) {
	inspect, err := c.Inspect(ctx)
	if err != nil {
		return "", err
	}
	if inspect.ContainerJSONBase.HostConfig.NetworkMode == "host" {
		return port, nil
	}

	ports := inspect.NetworkSettings.Ports

	for k, p := range ports {
		if k.Port() != port.Port() {
			continue
		}
		if port.Proto() != "" && k.Proto() != port.Proto() {
			continue
		}
		if len(p) == 0 {
			continue
		}
		return nat.NewPort(k.Proto(), p[0].HostPort)
	}

	return "", errors.New("port not found")
}

// NetworkAliases gets the aliases of the container for the networks it is attached to.
func (c *DockerContainer) NetworkAliases(ctx context.Context) (map[string][]string, error) {
	inspect, err := c.Inspect(ctx)
	if err != nil {
		return map[string][]string{}, err
	}

	networks := inspect.NetworkSettings.Networks

	a := map[string][]string{}

	for k := range networks {
		a[k] = networks[k].Aliases
	}

	return a, nil
}

// Networks gets the names of the networks the container is attached to.
func (c *DockerContainer) Networks(ctx context.Context) ([]string, error) {
	inspect, err := c.Inspect(ctx)
	if err != nil {
		return []string{}, err
	}

	networks := inspect.NetworkSettings.Networks

	n := []string{}

	for k := range networks {
		n = append(n, k)
	}

	return n, nil
}

// PortEndpoint gets proto://host:port string for the given exposed port
// Will returns just host:port if proto is ""
func (c *DockerContainer) PortEndpoint(ctx context.Context, port nat.Port, proto string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	outerPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}

	protoFull := ""
	if proto != "" {
		protoFull = fmt.Sprintf("%s://", proto)
	}

	return fmt.Sprintf("%s%s:%s", protoFull, host, outerPort.Port()), nil
}

// Printf prints a formatted string to the container logger.
func (c *DockerContainer) Printf(format string, args ...interface{}) {
	c.logger.Printf(format, args...)
}

// printLogs is a helper function that will print the logs of a Docker container
// We are going to use this helper function to inform the user of the logs when an error occurs
func (c *DockerContainer) printLogs(ctx context.Context, cause error) {
	reader, err := c.Logs(ctx)
	if err != nil {
		c.Printf("failed accessing container logs: %v\n", err)
		return
	}

	b, err := io.ReadAll(reader)
	if err != nil {
		c.logger.Printf("failed reading container logs: %v\n", err)
		return
	}

	c.Printf("container logs (%s):\n%s", cause, b)
}

// SessionID returns the session ID for the container
func (c *DockerContainer) SessionID() string {
	return c.sessionID
}

// SetLogger sets the logger for the container
// Used by Compose module
func (c *DockerContainer) SetLogger(logger log.Logging) {
	c.logger = logger
}

// SetTerminationSignal sets the termination signal for the container
// Used by Compose module
func (c *DockerContainer) SetTerminationSignal(signal chan bool) {
	c.terminationSignal = signal
}

// Start will start an already created container
func (c *DockerContainer) Start(ctx context.Context) error {
	err := c.startingHook(ctx)
	if err != nil {
		return err
	}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := cli.ContainerStart(ctx, c.ID, container.StartOptions{}); err != nil {
		return err
	}

	err = c.startedHook(ctx)
	if err != nil {
		return err
	}

	c.isRunning = true

	err = c.readiedHook(ctx)
	if err != nil {
		return err
	}

	return nil
}

// State returns container's running state. This method does not use the cache
// and always fetches the latest state from the Docker daemon.
func (c *DockerContainer) State(ctx context.Context) (*types.ContainerState, error) {
	inspect, err := c.inspectRawContainer(ctx)
	if err != nil {
		if c.raw != nil {
			return c.raw.State, err
		}
		return nil, err
	}
	return inspect.State, nil
}

// Stop will stop an already started container
//
// In case the container fails to stop
// gracefully within a time frame specified by the timeout argument,
// it is forcefully terminated (killed).
//
// If the timeout is nil, the container's StopTimeout value is used, if set,
// otherwise the engine default. A negative timeout value can be specified,
// meaning no timeout, i.e. no forceful termination is performed.
func (c *DockerContainer) Stop(ctx context.Context, timeout *time.Duration) error {
	err := c.stoppingHook(ctx)
	if err != nil {
		return err
	}

	var options container.StopOptions

	if timeout != nil {
		timeoutSeconds := int(timeout.Seconds())
		options.Timeout = &timeoutSeconds
	}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := cli.ContainerStop(ctx, c.ID, options); err != nil {
		return err
	}

	c.isRunning = false
	c.raw = nil // invalidate the cache, as the container representation will change after stopping

	err = c.stoppedHook(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Terminate is used to kill the container. It is usually triggered by as defer function.
func (c *DockerContainer) Terminate(ctx context.Context) error {
	select {
	// close reaper if it was created
	case c.terminationSignal <- true:
	default:
	}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()

	errs := []error{
		c.terminatingHook(ctx),
		cli.ContainerRemove(ctx, c.GetContainerID(), container.RemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}),
		c.terminatedHook(ctx),
	}

	if c.imageWasBuilt && !c.keepBuiltImage {
		_, err := cli.ImageRemove(ctx, c.Image, image.RemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if err != nil {
			errs = append(errs, err)
		} else {
			c.Printf("🧹 Successfully removed image %s", c.Image)
		}
	}

	c.sessionID = ""
	c.isRunning = false
	c.raw = nil // invalidate the cache here too

	return errors.Join(errs...)
}

func (c *DockerContainer) WaitUntilReady(ctx context.Context) error {
	if c.WaitingFor == nil {
		return nil
	}

	// if a Wait Strategy has been specified, wait before returning
	c.Printf(
		"⏳ Waiting for container id %s image: %s. Waiting for: %+v",
		c.GetContainerID()[:12], c.GetImage(), c.WaitingFor,
	)

	if err := c.WaitingFor.WaitUntilReady(ctx, c); err != nil {
		return err
	}

	c.isRunning = true

	return nil
}

// ------------------------------
// Container Lifecycle Hooks
// ------------------------------

// creatingHook is a hook that will be called before a container is created.
func (req Request) creatingHook(ctx context.Context) error {
	errs := make([]error, len(req.LifecycleHooks))
	for i, lifecycleHooks := range req.LifecycleHooks {
		errs[i] = lifecycleHooks.Creating(ctx)(&req)
	}

	return errors.Join(errs...)
}

// createdHook is a hook that will be called after a container is created.
func (c *DockerContainer) createdHook(ctx context.Context) error {
	return c.applyCreatedLifecycleHooks(ctx, false, func(lifecycleHooks LifecycleHooks) []CreatedContainerHook {
		return lifecycleHooks.PostCreates
	})
}

// startingHook is a hook that will be called before a container is started.
func (c *DockerContainer) startingHook(ctx context.Context) error {
	return c.applyCreatedLifecycleHooks(ctx, true, func(lifecycleHooks LifecycleHooks) []CreatedContainerHook {
		return lifecycleHooks.PreStarts
	})
}

// startedHook is a hook that will be called after a container is started.
func (c *DockerContainer) startedHook(ctx context.Context) error {
	return c.applyStartedLifecycleHooks(ctx, true, func(lifecycleHooks LifecycleHooks) []StartedContainerHook {
		return lifecycleHooks.PostStarts
	})
}

// readiedHook is a hook that will be called after a container is ready.
func (c *DockerContainer) readiedHook(ctx context.Context) error {
	return c.applyStartedLifecycleHooks(ctx, true, func(lifecycleHooks LifecycleHooks) []StartedContainerHook {
		return lifecycleHooks.PostReadies
	})
}

// stoppingHook is a hook that will be called before a container is stopped.
func (c *DockerContainer) stoppingHook(ctx context.Context) error {
	return c.applyStartedLifecycleHooks(ctx, false, func(lifecycleHooks LifecycleHooks) []StartedContainerHook {
		return lifecycleHooks.PreStops
	})
}

// stoppedHook is a hook that will be called after a container is stopped.
func (c *DockerContainer) stoppedHook(ctx context.Context) error {
	return c.applyStartedLifecycleHooks(ctx, false, func(lifecycleHooks LifecycleHooks) []StartedContainerHook {
		return lifecycleHooks.PostStops
	})
}

// terminatingHook is a hook that will be called before a container is terminated.
func (c *DockerContainer) terminatingHook(ctx context.Context) error {
	return c.applyStartedLifecycleHooks(ctx, false, func(lifecycleHooks LifecycleHooks) []StartedContainerHook {
		return lifecycleHooks.PreTerminates
	})
}

// terminatedHook is a hook that will be called after a container is terminated.
func (c *DockerContainer) terminatedHook(ctx context.Context) error {
	return c.applyStartedLifecycleHooks(ctx, false, func(lifecycleHooks LifecycleHooks) []StartedContainerHook {
		return lifecycleHooks.PostTerminates
	})
}
