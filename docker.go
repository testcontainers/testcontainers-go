package testcontainers

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/filters"

	"github.com/cenkalti/backoff/v4"
	"github.com/containerd/containerd/platforms"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/magiconair/properties"
	"github.com/moby/term"
	specs "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	// Implement interfaces
	_ Container = (*DockerContainer)(nil)

	ErrDuplicateMountTarget = errors.New("duplicate mount target detected")
)

const (
	Bridge        = "bridge" // Bridge network name (as well as driver)
	Podman        = "podman"
	ReaperDefault = "reaper_default" // Default network name when bridge is not available
)

// DockerContainer represents a container started using Docker
type DockerContainer struct {
	// Container ID from Docker
	ID         string
	WaitingFor wait.Strategy
	Image      string

	isRunning         bool
	imageWasBuilt     bool
	provider          *DockerProvider
	sessionID         uuid.UUID
	terminationSignal chan bool
	skipReaper        bool
	consumers         []LogConsumer
	raw               *types.ContainerJSON
	stopProducer      chan bool
	logger            Logging
}

func (c *DockerContainer) GetContainerID() string {
	return c.ID
}

func (c *DockerContainer) IsRunning() bool {
	return c.isRunning
}

// Endpoint gets proto://host:port string for the first exposed port
// Will returns just host:port if proto is ""
func (c *DockerContainer) Endpoint(ctx context.Context, proto string) (string, error) {
	ports, err := c.Ports(ctx)
	if err != nil {
		return "", err
	}

	// get first port
	var firstPort nat.Port
	for p := range ports {
		firstPort = p
		break
	}

	return c.PortEndpoint(ctx, firstPort, proto)
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

// Host gets host (ip or name) of the docker daemon where the container port is exposed
// Warning: this is based on your Docker host setting. Will fail if using an SSH tunnel
// You can use the "TC_HOST" env variable to set this yourself
func (c *DockerContainer) Host(ctx context.Context) (string, error) {
	host, err := c.provider.daemonHost(ctx)
	if err != nil {
		return "", err
	}
	return host, nil
}

// MappedPort gets externally mapped port for a container port
func (c *DockerContainer) MappedPort(ctx context.Context, port nat.Port) (nat.Port, error) {
	inspect, err := c.inspectContainer(ctx)
	if err != nil {
		return "", err
	}
	if inspect.ContainerJSONBase.HostConfig.NetworkMode == "host" {
		return port, nil
	}
	ports, err := c.Ports(ctx)
	if err != nil {
		return "", err
	}

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

// Ports gets the exposed ports for the container.
func (c *DockerContainer) Ports(ctx context.Context) (nat.PortMap, error) {
	inspect, err := c.inspectContainer(ctx)
	if err != nil {
		return nil, err
	}
	return inspect.NetworkSettings.Ports, nil
}

// SessionID gets the current session id
func (c *DockerContainer) SessionID() string {
	return c.sessionID.String()
}

// Start will start an already created container
func (c *DockerContainer) Start(ctx context.Context) error {
	shortID := c.ID[:12]
	c.logger.Printf("Starting container id: %s image: %s", shortID, c.Image)

	if err := c.provider.client.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// if a Wait Strategy has been specified, wait before returning
	if c.WaitingFor != nil {
		c.logger.Printf("Waiting for container id %s image: %s", shortID, c.Image)
		if err := c.WaitingFor.WaitUntilReady(ctx, c); err != nil {
			return err
		}
	}
	c.logger.Printf("Container is ready id: %s image: %s", shortID, c.Image)
	c.isRunning = true
	return nil
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
	shortID := c.ID[:12]
	c.logger.Printf("Stopping container id: %s image: %s", shortID, c.Image)

	if err := c.provider.client.ContainerStop(ctx, c.ID, timeout); err != nil {
		return err
	}

	c.logger.Printf("Container is stopped id: %s image: %s", shortID, c.Image)
	c.isRunning = false
	return nil
}

// Terminate is used to kill the container. It is usually triggered by as defer function.
func (c *DockerContainer) Terminate(ctx context.Context) error {
	select {
	// close reaper if it was created
	case c.terminationSignal <- true:
	default:
	}
	err := c.provider.client.ContainerRemove(ctx, c.GetContainerID(), types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return err
	}

	if c.imageWasBuilt {
		_, err := c.provider.client.ImageRemove(ctx, c.Image, types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if err != nil {
			return err
		}
	}

	if err := c.provider.client.Close(); err != nil {
		return err
	}

	c.sessionID = uuid.UUID{}
	c.isRunning = false
	return nil
}

// update container raw info
func (c *DockerContainer) inspectRawContainer(ctx context.Context) (*types.ContainerJSON, error) {
	inspect, err := c.provider.client.ContainerInspect(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.raw = &inspect
	return c.raw, nil
}

func (c *DockerContainer) inspectContainer(ctx context.Context) (*types.ContainerJSON, error) {
	inspect, err := c.provider.client.ContainerInspect(ctx, c.ID)
	if err != nil {
		return nil, err
	}

	return &inspect, nil
}

// Logs will fetch both STDOUT and STDERR from the current container. Returns a
// ReadCloser and leaves it up to the caller to extract what it wants.
func (c *DockerContainer) Logs(ctx context.Context) (io.ReadCloser, error) {

	const streamHeaderSize = 8

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}

	rc, err := c.provider.client.ContainerLogs(ctx, c.ID, options)
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	r := bufio.NewReader(rc)

	go func() {
		var (
			isPrefix    = true
			lineStarted = true
			line        []byte
		)
		for err == nil {
			line, isPrefix, err = r.ReadLine()

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

// FollowOutput adds a LogConsumer to be sent logs from the container's
// STDOUT and STDERR
func (c *DockerContainer) FollowOutput(consumer LogConsumer) {
	if c.consumers == nil {
		c.consumers = []LogConsumer{
			consumer,
		}
	} else {
		c.consumers = append(c.consumers, consumer)
	}
}

// Name gets the name of the container.
func (c *DockerContainer) Name(ctx context.Context) (string, error) {
	inspect, err := c.inspectContainer(ctx)
	if err != nil {
		return "", err
	}
	return inspect.Name, nil
}

// State returns container's running state
func (c *DockerContainer) State(ctx context.Context) (*types.ContainerState, error) {
	inspect, err := c.inspectRawContainer(ctx)
	if err != nil {
		return c.raw.State, err
	}
	return inspect.State, nil
}

// Networks gets the names of the networks the container is attached to.
func (c *DockerContainer) Networks(ctx context.Context) ([]string, error) {
	inspect, err := c.inspectContainer(ctx)
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

// ContainerIP gets the IP address of the primary network within the container.
func (c *DockerContainer) ContainerIP(ctx context.Context) (string, error) {
	inspect, err := c.inspectContainer(ctx)
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

// NetworkAliases gets the aliases of the container for the networks it is attached to.
func (c *DockerContainer) NetworkAliases(ctx context.Context) (map[string][]string, error) {
	inspect, err := c.inspectContainer(ctx)
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

func (c *DockerContainer) Exec(ctx context.Context, cmd []string) (int, io.Reader, error) {
	cli := c.provider.client
	response, err := cli.ContainerExecCreate(ctx, c.ID, types.ExecConfig{
		Cmd:          cmd,
		Detach:       false,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return 0, nil, err
	}

	hijack, err := cli.ContainerExecAttach(ctx, response.ID, types.ExecStartCheck{})
	if err != nil {
		return 0, nil, err
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

	return exitCode, hijack.Reader, nil
}

type FileFromContainer struct {
	underlying *io.ReadCloser
	tarreader  *tar.Reader
}

func (fc *FileFromContainer) Read(b []byte) (int, error) {
	return (*fc.tarreader).Read(b)
}

func (fc *FileFromContainer) Close() error {
	return (*fc.underlying).Close()
}

func (c *DockerContainer) CopyFileFromContainer(ctx context.Context, filePath string) (io.ReadCloser, error) {
	r, _, err := c.provider.client.CopyFromContainer(ctx, c.ID, filePath)
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
		underlying: &r,
		tarreader:  tarReader,
	}

	return ret, nil
}

// CopyDirToContainer copies the contents of a directory to a parent path in the container. This parent path must exist in the container first
// as we cannot create it
func (c *DockerContainer) CopyDirToContainer(ctx context.Context, hostDirPath string, containerParentPath string, fileMode int64) error {
	dir, err := isDir(hostDirPath)
	if err != nil {
		return err
	}

	if !dir {
		// it's not a dir: let the consumer to handle an error
		return fmt.Errorf("path %s is not a directory", hostDirPath)
	}

	buff, err := tarDir(hostDirPath, fileMode)
	if err != nil {
		return err
	}

	// create the directory under its parent
	parent := filepath.Dir(containerParentPath)

	return c.provider.client.CopyToContainer(ctx, c.ID, parent, buff, types.CopyToContainerOptions{})
}

func (c *DockerContainer) CopyFileToContainer(ctx context.Context, hostFilePath string, containerFilePath string, fileMode int64) error {
	dir, err := isDir(hostFilePath)
	if err != nil {
		return err
	}

	if dir {
		return c.CopyDirToContainer(ctx, hostFilePath, containerFilePath, fileMode)
	}

	fileContent, err := ioutil.ReadFile(hostFilePath)
	if err != nil {
		return err
	}
	return c.CopyToContainer(ctx, fileContent, containerFilePath, fileMode)
}

// CopyToContainer copies fileContent data to a file in container
func (c *DockerContainer) CopyToContainer(ctx context.Context, fileContent []byte, containerFilePath string, fileMode int64) error {
	buffer, err := tarFile(fileContent, containerFilePath, fileMode)
	if err != nil {
		return err
	}

	return c.provider.client.CopyToContainer(ctx, c.ID, filepath.Dir(containerFilePath), buffer, types.CopyToContainerOptions{})
}

// StartLogProducer will start a concurrent process that will continuously read logs
// from the container and will send them to each added LogConsumer
func (c *DockerContainer) StartLogProducer(ctx context.Context) error {
	go func() {
		since := ""
		// if the socket is closed we will make additional logs request with updated Since timestamp
	BEGIN:
		options := types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Since:      since,
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		r, err := c.provider.client.ContainerLogs(ctx, c.GetContainerID(), options)
		if err != nil {
			// if we can't get the logs, panic, we can't return an error to anything
			// from within this goroutine
			panic(err)
		}

		for {
			select {
			case <-c.stopProducer:
				err := r.Close()
				if err != nil {
					// we can't close the read closer, this should never happen
					panic(err)
				}
				return
			default:
				h := make([]byte, 8)
				_, err := r.Read(h)
				if err != nil {
					// proper type matching requires https://go-review.googlesource.com/c/go/+/250357/ (go 1.16)
					if strings.Contains(err.Error(), "use of closed network connection") {
						now := time.Now()
						since = fmt.Sprintf("%d.%09d", now.Unix(), int64(now.Nanosecond()))
						goto BEGIN
					}
					// this explicitly ignores errors
					// because we want to keep procesing even if one of our reads fails
					continue
				}

				count := binary.BigEndian.Uint32(h[4:])
				if count == 0 {
					continue
				}
				logType := h[0]
				if logType > 2 {
					_, _ = fmt.Fprintf(os.Stderr, "received invalid log type: %d", logType)
					// sometimes docker returns logType = 3 which is an undocumented log type, so treat it as stdout
					logType = 1
				}

				// a map of the log type --> int representation in the header, notice the first is blank, this is stdin, but the go docker client doesn't allow following that in logs
				logTypes := []string{"", StdoutLog, StderrLog}

				b := make([]byte, count)
				_, err = r.Read(b)
				if err != nil {
					// TODO: add-logger: use logger to log out this error
					_, _ = fmt.Fprintf(os.Stderr, "error occurred reading log with known length %s", err.Error())
					continue
				}
				for _, c := range c.consumers {
					c.Accept(Log{
						LogType: logTypes[logType],
						Content: b,
					})
				}
			}
		}
	}()

	return nil
}

// StopLogProducer will stop the concurrent process that is reading logs
// and sending them to each added LogConsumer
func (c *DockerContainer) StopLogProducer() error {
	c.stopProducer <- true
	return nil
}

// DockerNetwork represents a network started using Docker
type DockerNetwork struct {
	ID                string // Network ID from Docker
	Driver            string
	Name              string
	provider          *DockerProvider
	terminationSignal chan bool
}

// Remove is used to remove the network. It is usually triggered by as defer function.
func (n *DockerNetwork) Remove(ctx context.Context) error {
	select {
	// close reaper if it was created
	case n.terminationSignal <- true:
	default:
	}
	return n.provider.client.NetworkRemove(ctx, n.ID)
}

// DockerProvider implements the ContainerProvider interface
type DockerProvider struct {
	*DockerProviderOptions
	client    *client.Client
	host      string
	hostCache string
	config    TestContainersConfig
}

var _ ContainerProvider = (*DockerProvider)(nil)

// or through Decode
type TestContainersConfig struct {
	Host           string `properties:"docker.host,default="`
	TLSVerify      int    `properties:"docker.tls.verify,default=0"`
	CertPath       string `properties:"docker.cert.path,default="`
	RyukPrivileged bool   `properties:"ryuk.container.privileged,default=false"`
}

type (
	// DockerProviderOptions defines options applicable to DockerProvider
	DockerProviderOptions struct {
		defaultBridgeNetworkName string
		*GenericProviderOptions
	}

	// DockerProviderOption defines a common interface to modify DockerProviderOptions
	// These can be passed to NewDockerProvider in a variadic way to customize the returned DockerProvider instance
	DockerProviderOption interface {
		ApplyDockerTo(opts *DockerProviderOptions)
	}

	// DockerProviderOptionFunc is a shorthand to implement the DockerProviderOption interface
	DockerProviderOptionFunc func(opts *DockerProviderOptions)
)

func (f DockerProviderOptionFunc) ApplyDockerTo(opts *DockerProviderOptions) {
	f(opts)
}

func Generic2DockerOptions(opts ...GenericProviderOption) []DockerProviderOption {
	converted := make([]DockerProviderOption, 0, len(opts))
	for _, o := range opts {
		switch c := o.(type) {
		case DockerProviderOption:
			converted = append(converted, c)
		default:
			converted = append(converted, DockerProviderOptionFunc(func(opts *DockerProviderOptions) {
				o.ApplyGenericTo(opts.GenericProviderOptions)
			}))
		}
	}

	return converted
}

func WithDefaultBridgeNetwork(bridgeNetworkName string) DockerProviderOption {
	return DockerProviderOptionFunc(func(opts *DockerProviderOptions) {
		opts.defaultBridgeNetworkName = bridgeNetworkName
	})
}

func NewDockerClient() (cli *client.Client, host string, tcConfig TestContainersConfig, err error) {
	tcConfig = configureTC()

	host = tcConfig.Host

	opts := []client.Opt{client.FromEnv}
	if host != "" {
		opts = append(opts, client.WithHost(host))

		// for further informacion, read https://docs.docker.com/engine/security/protect-access/
		if tcConfig.TLSVerify == 1 {
			cacertPath := filepath.Join(tcConfig.CertPath, "ca.pem")
			certPath := filepath.Join(tcConfig.CertPath, "cert.pem")
			keyPath := filepath.Join(tcConfig.CertPath, "key.pem")

			opts = append(opts, client.WithTLSClientConfig(cacertPath, certPath, keyPath))
		}
	} else if dockerHostEnv := os.Getenv("DOCKER_HOST"); dockerHostEnv != "" {
		host = dockerHostEnv
	} else {
		host = "unix:///var/run/docker.sock"
	}

	cli, err = client.NewClientWithOpts(opts...)

	if err != nil {
		return nil, "", TestContainersConfig{}, err
	}

	cli.NegotiateAPIVersion(context.Background())

	return cli, host, tcConfig, nil
}

// NewDockerProvider creates a Docker provider with the EnvClient
func NewDockerProvider(provOpts ...DockerProviderOption) (*DockerProvider, error) {
	o := &DockerProviderOptions{
		GenericProviderOptions: &GenericProviderOptions{
			Logger: Logger,
		},
	}

	for idx := range provOpts {
		provOpts[idx].ApplyDockerTo(o)
	}

	c, host, tcConfig, err := NewDockerClient()
	if err != nil {
		return nil, err
	}

	_, err = c.Ping(context.TODO())
	if err != nil {
		// fallback to environment
		c, err = client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			return nil, err
		}
	}

	c.NegotiateAPIVersion(context.Background())
	p := &DockerProvider{
		DockerProviderOptions: o,
		host:                  host,
		client:                c,
		config:                tcConfig,
	}

	return p, nil
}

// configureTC reads from testcontainers properties file, if it exists
// it is possible that certain values get overridden when set as environment variables
func configureTC() TestContainersConfig {
	config := TestContainersConfig{}

	applyEnvironmentConfiguration := func(config TestContainersConfig) TestContainersConfig {
		ryukPrivilegedEnv := os.Getenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED")
		if ryukPrivilegedEnv != "" {
			config.RyukPrivileged = ryukPrivilegedEnv == "true"
		}

		return config
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return applyEnvironmentConfiguration(config)
	}

	tcProp := filepath.Join(home, ".testcontainers.properties")
	// init from a file
	properties, err := properties.LoadFile(tcProp, properties.UTF8)
	if err != nil {
		return applyEnvironmentConfiguration(config)
	}

	if err := properties.Decode(&config); err != nil {
		fmt.Printf("invalid testcontainers properties file, returning an empty Testcontainers configuration: %v\n", err)
		return applyEnvironmentConfiguration(config)
	}

	return applyEnvironmentConfiguration(config)
}

// BuildImage will build and image from context and Dockerfile, then return the tag
func (p *DockerProvider) BuildImage(ctx context.Context, img ImageBuildInfo) (string, error) {
	repo := uuid.New()
	tag := uuid.New()

	repoTag := fmt.Sprintf("%s:%s", repo, tag)

	buildContext, err := img.GetContext()
	if err != nil {
		return "", err
	}

	buildOptions := types.ImageBuildOptions{
		BuildArgs:   img.GetBuildArgs(),
		Dockerfile:  img.GetDockerfile(),
		Context:     buildContext,
		Tags:        []string{repoTag},
		Remove:      true,
		ForceRemove: true,
	}

	resp, err := p.client.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return "", err
	}

	if img.ShouldPrintBuildLog() {
		termFd, isTerm := term.GetFdInfo(os.Stderr)
		err = jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil)
		if err != nil {
			return "", err
		}
	}

	// need to read the response from Docker, I think otherwise the image
	// might not finish building before continuing to execute here
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	_ = resp.Body.Close()

	return repoTag, nil
}

// CreateContainer fulfills a request for a container without starting it
func (p *DockerProvider) CreateContainer(ctx context.Context, req ContainerRequest) (Container, error) {
	var err error

	// Make sure that bridge network exists
	// In case it is disabled we will create reaper_default network
	if p.DefaultNetwork == "" {
		p.DefaultNetwork, err = p.getDefaultNetwork(ctx, p.client)
		if err != nil {
			return nil, err
		}
	}

	// If default network is not bridge make sure it is attached to the request
	// as container won't be attached to it automatically
	// in case of Podman the bridge network is called 'podman' as 'bridge' would conflict
	if p.DefaultNetwork != p.defaultBridgeNetworkName {
		isAttached := false
		for _, net := range req.Networks {
			if net == p.DefaultNetwork {
				isAttached = true
				break
			}
		}

		if !isAttached {
			req.Networks = append(req.Networks, p.DefaultNetwork)
		}
	}

	env := []string{}
	for envKey, envVar := range req.Env {
		env = append(env, envKey+"="+envVar)
	}

	if req.Labels == nil {
		req.Labels = make(map[string]string)
	}

	sessionID := uuid.New()

	var termSignal chan bool
	if !req.SkipReaper {
		r, err := NewReaper(context.WithValue(ctx, dockerHostContextKey, p.host), sessionID.String(), p, req.ReaperImage)
		if err != nil {
			return nil, fmt.Errorf("%w: creating reaper failed", err)
		}
		termSignal, err = r.Connect()
		if err != nil {
			return nil, fmt.Errorf("%w: connecting to reaper failed", err)
		}
		for k, v := range r.Labels() {
			if _, ok := req.Labels[k]; !ok {
				req.Labels[k] = v
			}
		}
	}

	if err = req.Validate(); err != nil {
		return nil, err
	}

	var tag string
	var platform *specs.Platform

	if req.ShouldBuildImage() {
		tag, err = p.BuildImage(ctx, &req)
		if err != nil {
			return nil, err
		}
	} else {
		tag = req.Image

		if req.ImagePlatform != "" {
			p, err := platforms.Parse(req.ImagePlatform)
			if err != nil {
				return nil, fmt.Errorf("invalid platform %s: %w", req.ImagePlatform, err)
			}
			platform = &p
		}

		var shouldPullImage bool

		if req.AlwaysPullImage {
			shouldPullImage = true // If requested always attempt to pull image
		} else {
			image, _, err := p.client.ImageInspectWithRaw(ctx, tag)
			if err != nil {
				if client.IsErrNotFound(err) {
					shouldPullImage = true
				} else {
					return nil, err
				}
			}
			if platform != nil && (image.Architecture != platform.Architecture || image.Os != platform.OS) {
				shouldPullImage = true
			}
		}

		if shouldPullImage {
			pullOpt := types.ImagePullOptions{
				Platform: req.ImagePlatform, // may be empty
			}

			if req.RegistryCred != "" {
				pullOpt.RegistryAuth = req.RegistryCred
			}

			if err := p.attemptToPullImage(ctx, tag, pullOpt); err != nil {
				return nil, err
			}
		}
	}

	exposedPorts := req.ExposedPorts
	if len(exposedPorts) == 0 {
		image, _, err := p.client.ImageInspectWithRaw(ctx, tag)
		if err != nil {
			return nil, err
		}
		for p, _ := range image.ContainerConfig.ExposedPorts {
			exposedPorts = append(exposedPorts, string(p))
		}
	}

	exposedPortSet, exposedPortMap, err := nat.ParsePortSpecs(exposedPorts)
	if err != nil {
		return nil, err
	}

	dockerInput := &container.Config{
		Entrypoint:   req.Entrypoint,
		Image:        tag,
		Env:          env,
		ExposedPorts: exposedPortSet,
		Labels:       req.Labels,
		Cmd:          req.Cmd,
		Hostname:     req.Hostname,
		User:         req.User,
	}

	// prepare mounts
	mounts := mapToDockerMounts(req.Mounts)

	hostConfig := &container.HostConfig{
		ExtraHosts:   req.ExtraHosts,
		PortBindings: exposedPortMap,
		Binds:        req.Binds,
		Mounts:       mounts,
		Tmpfs:        req.Tmpfs,
		AutoRemove:   req.AutoRemove,
		Privileged:   req.Privileged,
		NetworkMode:  req.NetworkMode,
		Resources:    req.Resources,
		ShmSize:      req.ShmSize,
	}

	endpointConfigs := map[string]*network.EndpointSettings{}

	// #248: Docker allows only one network to be specified during container creation
	// If there is more than one network specified in the request container should be attached to them
	// once it is created. We will take a first network if any specified in the request and use it to create container
	if len(req.Networks) > 0 {
		attachContainerTo := req.Networks[0]

		nw, err := p.GetNetwork(ctx, NetworkRequest{
			Name: attachContainerTo,
		})
		if err == nil {
			endpointSetting := network.EndpointSettings{
				Aliases:   req.NetworkAliases[attachContainerTo],
				NetworkID: nw.ID,
			}
			endpointConfigs[attachContainerTo] = &endpointSetting
		}
	}

	networkingConfig := network.NetworkingConfig{
		EndpointsConfig: endpointConfigs,
	}

	resp, err := p.client.ContainerCreate(ctx, dockerInput, hostConfig, &networkingConfig, platform, req.Name)
	if err != nil {
		return nil, err
	}

	// #248: If there is more than one network specified in the request attach newly created container to them one by one
	if len(req.Networks) > 1 {
		for _, n := range req.Networks[1:] {
			nw, err := p.GetNetwork(ctx, NetworkRequest{
				Name: n,
			})
			if err == nil {
				endpointSetting := network.EndpointSettings{
					Aliases: req.NetworkAliases[n],
				}
				err = p.client.NetworkConnect(ctx, nw.ID, resp.ID, &endpointSetting)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	c := &DockerContainer{
		ID:                resp.ID,
		WaitingFor:        req.WaitingFor,
		Image:             tag,
		imageWasBuilt:     req.ShouldBuildImage(),
		sessionID:         sessionID,
		provider:          p,
		terminationSignal: termSignal,
		skipReaper:        req.SkipReaper,
		stopProducer:      make(chan bool),
		logger:            p.Logger,
	}

	for _, f := range req.Files {
		err := c.CopyFileToContainer(ctx, f.HostFilePath, f.ContainerFilePath, f.FileMode)
		if err != nil {
			return nil, fmt.Errorf("can't copy %s to container: %w", f.HostFilePath, err)
		}
	}

	return c, nil
}

func (p *DockerProvider) findContainerByName(ctx context.Context, name string) (*types.Container, error) {
	if name == "" {
		return nil, nil
	}
	filter := filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: name,
	})
	containers, err := p.client.ContainerList(ctx, types.ContainerListOptions{Filters: filter})
	if err != nil {
		return nil, err
	}
	if len(containers) > 0 {
		return &containers[0], nil
	}
	return nil, nil
}

func (p *DockerProvider) ReuseOrCreateContainer(ctx context.Context, req ContainerRequest) (Container, error) {
	c, err := p.findContainerByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return p.CreateContainer(ctx, req)
	}

	sessionID := uuid.New()
	var termSignal chan bool
	if !req.SkipReaper {
		r, err := NewReaper(ctx, sessionID.String(), p, req.ReaperImage)
		if err != nil {
			return nil, fmt.Errorf("%w: creating reaper failed", err)
		}
		termSignal, err = r.Connect()
		if err != nil {
			return nil, fmt.Errorf("%w: connecting to reaper failed", err)
		}
	}
	dc := &DockerContainer{
		ID:                c.ID,
		WaitingFor:        req.WaitingFor,
		Image:             c.Image,
		sessionID:         sessionID,
		provider:          p,
		terminationSignal: termSignal,
		skipReaper:        req.SkipReaper,
		stopProducer:      make(chan bool),
		logger:            p.Logger,
		isRunning:         c.State == "running",
	}
	return dc, nil

}

// attemptToPullImage tries to pull the image while respecting the ctx cancellations.
// Besides, if the image cannot be pulled due to ErrorNotFound then no need to retry but terminate immediately.
func (p *DockerProvider) attemptToPullImage(ctx context.Context, tag string, pullOpt types.ImagePullOptions) error {
	var (
		err  error
		pull io.ReadCloser
	)
	err = backoff.Retry(func() error {
		pull, err = p.client.ImagePull(ctx, tag, pullOpt)
		if err != nil {
			if _, ok := err.(errdefs.ErrNotFound); ok {
				return backoff.Permanent(err)
			}
			Logger.Printf("Failed to pull image: %s, will retry", err)
			return err
		}
		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
	if err != nil {
		return err
	}
	defer pull.Close()

	// download of docker image finishes at EOF of the pull request
	_, err = ioutil.ReadAll(pull)
	return err
}

// Health measure the healthiness of the provider. Right now we leverage the
// docker-client ping endpoint to see if the daemon is reachable.
func (p *DockerProvider) Health(ctx context.Context) (err error) {
	_, err = p.client.Ping(ctx)
	return
}

// RunContainer takes a RequestContainer as input and it runs a container via the docker sdk
func (p *DockerProvider) RunContainer(ctx context.Context, req ContainerRequest) (Container, error) {
	c, err := p.CreateContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := c.Start(ctx); err != nil {
		return c, fmt.Errorf("%w: could not start container", err)
	}

	return c, nil
}

// Config provides the TestContainersConfig read from $HOME/.testcontainers.properties or
// the environment variables
func (p *DockerProvider) Config() TestContainersConfig {
	return p.config
}

// daemonHost gets the host or ip of the Docker daemon where ports are exposed on
// Warning: this is based on your Docker host setting. Will fail if using an SSH tunnel
// You can use the "TC_HOST" env variable to set this yourself
func (p *DockerProvider) daemonHost(ctx context.Context) (string, error) {
	if p.hostCache != "" {
		return p.hostCache, nil
	}

	host, exists := os.LookupEnv("TC_HOST")
	if exists {
		p.hostCache = host
		return p.hostCache, nil
	}

	// infer from Docker host
	url, err := url.Parse(p.client.DaemonHost())
	if err != nil {
		return "", err
	}

	switch url.Scheme {
	case "http", "https", "tcp":
		p.hostCache = url.Hostname()
	case "unix", "npipe":
		if inAContainer() {
			ip, err := p.GetGatewayIP(ctx)
			if err != nil {
				// fallback to getDefaultGatewayIP
				ip, err = getDefaultGatewayIP()
				if err != nil {
					ip = "localhost"
				}
			}
			p.hostCache = ip
		} else {
			p.hostCache = "localhost"
		}
	default:
		return "", errors.New("Could not determine host through env or docker host")
	}

	return p.hostCache, nil
}

// CreateNetwork returns the object representing a new network identified by its name
func (p *DockerProvider) CreateNetwork(ctx context.Context, req NetworkRequest) (Network, error) {
	var err error

	// Make sure that bridge network exists
	// In case it is disabled we will create reaper_default network
	if p.DefaultNetwork == "" {
		if p.DefaultNetwork, err = p.getDefaultNetwork(ctx, p.client); err != nil {
			return nil, err
		}
	}

	if req.Labels == nil {
		req.Labels = make(map[string]string)
	}

	nc := types.NetworkCreate{
		Driver:         req.Driver,
		CheckDuplicate: req.CheckDuplicate,
		Internal:       req.Internal,
		EnableIPv6:     req.EnableIPv6,
		Attachable:     req.Attachable,
		Labels:         req.Labels,
		IPAM:           req.IPAM,
	}

	sessionID := uuid.New()

	var termSignal chan bool
	if !req.SkipReaper {
		r, err := NewReaper(context.WithValue(ctx, dockerHostContextKey, p.host), sessionID.String(), p, req.ReaperImage)
		if err != nil {
			return nil, fmt.Errorf("%w: creating network reaper failed", err)
		}
		termSignal, err = r.Connect()
		if err != nil {
			return nil, fmt.Errorf("%w: connecting to network reaper failed", err)
		}
		for k, v := range r.Labels() {
			if _, ok := req.Labels[k]; !ok {
				req.Labels[k] = v
			}
		}
	}

	response, err := p.client.NetworkCreate(ctx, req.Name, nc)
	if err != nil {
		return &DockerNetwork{}, err
	}

	n := &DockerNetwork{
		ID:                response.ID,
		Driver:            req.Driver,
		Name:              req.Name,
		terminationSignal: termSignal,
		provider:          p,
	}

	return n, nil
}

// GetNetwork returns the object representing the network identified by its name
func (p *DockerProvider) GetNetwork(ctx context.Context, req NetworkRequest) (types.NetworkResource, error) {
	networkResource, err := p.client.NetworkInspect(ctx, req.Name, types.NetworkInspectOptions{
		Verbose: true,
	})
	if err != nil {
		return types.NetworkResource{}, err
	}

	return networkResource, err
}

func (p *DockerProvider) GetGatewayIP(ctx context.Context) (string, error) {
	// Use a default network as defined in the DockerProvider
	if p.DefaultNetwork == "" {
		var err error
		p.DefaultNetwork, err = p.getDefaultNetwork(ctx, p.client)
		if err != nil {
			return "", err
		}
	}
	nw, err := p.GetNetwork(ctx, NetworkRequest{Name: p.DefaultNetwork})
	if err != nil {
		return "", err
	}

	var ip string
	for _, config := range nw.IPAM.Config {
		if config.Gateway != "" {
			ip = config.Gateway
			break
		}
	}
	if ip == "" {
		return "", errors.New("Failed to get gateway IP from network settings")
	}

	return ip, nil
}

func inAContainer() bool {
	// see https://github.com/testcontainers/testcontainers-java/blob/3ad8d80e2484864e554744a4800a81f6b7982168/core/src/main/java/org/testcontainers/dockerclient/DockerClientConfigUtils.java#L15
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

// deprecated
// see https://github.com/testcontainers/testcontainers-java/blob/main/core/src/main/java/org/testcontainers/dockerclient/DockerClientConfigUtils.java#L46
func getDefaultGatewayIP() (string, error) {
	// see https://github.com/testcontainers/testcontainers-java/blob/3ad8d80e2484864e554744a4800a81f6b7982168/core/src/main/java/org/testcontainers/dockerclient/DockerClientConfigUtils.java#L27
	cmd := exec.Command("sh", "-c", "ip route|awk '/default/ { print $3 }'")
	stdout, err := cmd.Output()
	if err != nil {
		return "", errors.New("Failed to detect docker host")
	}
	ip := strings.TrimSpace(string(stdout))
	if len(ip) == 0 {
		return "", errors.New("Failed to parse default gateway IP")
	}
	return ip, nil
}

func (p *DockerProvider) getDefaultNetwork(ctx context.Context, cli *client.Client) (string, error) {
	// Get list of available networks
	networkResources, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return "", err
	}

	reaperNetwork := ReaperDefault

	reaperNetworkExists := false

	for _, net := range networkResources {
		if net.Name == p.defaultBridgeNetworkName {
			return p.defaultBridgeNetworkName, nil
		}

		if net.Name == reaperNetwork {
			reaperNetworkExists = true
		}
	}

	// Create a bridge network for the container communications
	if !reaperNetworkExists {
		_, err = cli.NetworkCreate(ctx, reaperNetwork, types.NetworkCreate{
			Driver:     Bridge,
			Attachable: true,
			Labels: map[string]string{
				TestcontainerLabel: "true",
			},
		})

		if err != nil {
			return "", err
		}
	}

	return reaperNetwork, nil
}
