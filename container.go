package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/cenkalti/backoff/v4"
	"github.com/containerd/platforms"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	specs "github.com/opencontainers/image-spec/specs-go/v1"

	tcimage "github.com/testcontainers/testcontainers-go/image"
	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
	corecontainer "github.com/testcontainers/testcontainers-go/internal/core/container"
	corenetwork "github.com/testcontainers/testcontainers-go/internal/core/network"
	"github.com/testcontainers/testcontainers-go/internal/core/reaper"
	tclog "github.com/testcontainers/testcontainers-go/log"
)

var (
	reuseContainerMx  sync.Mutex
	ErrReuseEmptyName = errors.New("with reuse option a container name mustn't be empty")
)

func findContainerByName(ctx context.Context, name string) (*types.Container, error) {
	if name == "" {
		return nil, nil
	}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	// Note that, 'name' filter will use regex to find the containers
	filter := filters.NewArgs(filters.Arg("name", fmt.Sprintf("^%s$", name)))
	containers, err := cli.ContainerList(ctx, container.ListOptions{Filters: filter})
	if err != nil {
		return nil, err
	}

	if len(containers) > 0 {
		return &containers[0], nil
	}

	return nil, nil
}

func Run(ctx context.Context, req Request) (*DockerContainer, error) {
	if req.Reuse && req.Name == "" {
		return nil, ErrReuseEmptyName
	}

	if req.Logger == nil {
		req.Logger = tclog.StandardLogger()
	}

	var err error
	var c *DockerContainer
	if req.Reuse {
		// we must protect the reusability of the container in the case it's invoked
		// in a parallel execution, via ParallelContainers or t.Parallel()
		reuseContainerMx.Lock()
		defer reuseContainerMx.Unlock()

		c, err = reuseOrCreateContainer(ctx, req)
	} else {
		c, err = newContainer(ctx, req)
	}
	if err != nil {
		// At this point `c` might not be nil. Give the caller an opportunity to call Destroy on the container.
		return c, fmt.Errorf("%w: failed to create container", err)
	}

	if req.Started && !c.IsRunning() {
		if err := c.Start(ctx); err != nil {
			return c, fmt.Errorf("failed to start container: %w", err)
		}
	}
	return c, nil
}

func newContainer(ctx context.Context, req Request) (*DockerContainer, error) {
	if req.Logger == nil {
		req.Logger = tclog.StandardLogger()
	}

	// Make sure that bridge network exists
	// In case it is disabled we will create reaper_default network
	defaultNetwork, err := corenetwork.GetDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get default network: %w", err)
	}

	// If default network is not bridge make sure it is attached to the request
	// as container won't be attached to it automatically
	if defaultNetwork != corenetwork.Bridge {
		isAttached := false
		for _, net := range req.Networks {
			if net == defaultNetwork {
				isAttached = true
				break
			}
		}

		if !isAttached {
			req.Networks = append(req.Networks, defaultNetwork)
		}
	}

	imageName := req.Image

	env := []string{}
	for envKey, envVar := range req.Env {
		env = append(env, envKey+"="+envVar)
	}

	if req.Labels == nil {
		req.Labels = make(map[string]string)
	}

	tcConfig := config.Read()

	var termSignal chan bool
	// the reaper does not need to start a reaper for itself
	isReaperContainer := strings.HasSuffix(imageName, config.ReaperDefaultImage)
	if !tcConfig.RyukDisabled && !isReaperContainer {
		_, err := NewReaper(context.Background(), core.SessionID())
		if err != nil {
			return nil, fmt.Errorf("failed to create reaper: %w", err)
		}

		termSignal, err = reaper.Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to reaper: %w", err)
		}
	}

	// Cleanup on error, otherwise set termSignal to nil before successful return.
	defer func() {
		if termSignal != nil {
			termSignal <- true
		}
	}()

	if err = req.Validate(); err != nil {
		return nil, err
	}

	// always append the hub substitutor after the user-defined ones
	req.ImageSubstitutors = append(req.ImageSubstitutors, tcimage.NewPrependHubRegistry(tcConfig.HubImageNamePrefix))

	cli, err := core.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a Docker client: %w", err)
	}
	defer cli.Close()

	var platform *specs.Platform

	if req.ShouldBuildImage() {
		imageName, err = tcimage.Build(ctx, &req)
		if err != nil {
			return nil, err
		}
	} else {
		for _, is := range req.ImageSubstitutors {
			modifiedTag, err := is.Substitute(imageName)
			if err != nil {
				return nil, fmt.Errorf("failed to substitute image %s with %s: %w", imageName, is.Description(), err)
			}

			if modifiedTag != imageName {
				req.Logger.Printf("✍🏼 Replacing image with %s. From: %s to %s\n", is.Description(), imageName, modifiedTag)
				imageName = modifiedTag
			}
		}

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
			img, _, err := cli.ImageInspectWithRaw(ctx, imageName)
			if err != nil {
				if client.IsErrNotFound(err) {
					shouldPullImage = true
				} else {
					return nil, err
				}
			}
			if platform != nil && (img.Architecture != platform.Architecture || img.Os != platform.OS) {
				shouldPullImage = true
			}
		}

		if shouldPullImage {
			pullOpt := image.PullOptions{
				Platform: req.ImagePlatform, // may be empty
			}
			if err := tcimage.Pull(ctx, imageName, req.Logger, pullOpt); err != nil {
				return nil, err
			}
		}
	}

	if !isReaperContainer {
		// add the labels that the reaper will use to terminate the container to the request
		for k, v := range core.DefaultLabels(core.SessionID()) {
			req.Labels[k] = v
		}
	}

	dockerInput := &container.Config{
		Entrypoint: req.Entrypoint,
		Image:      imageName,
		Env:        env,
		Labels:     req.Labels,
		Cmd:        req.Cmd,
		Hostname:   req.Hostname,
		User:       req.User,
		WorkingDir: req.WorkingDir,
	}

	hostConfig := &container.HostConfig{
		Privileged: req.Privileged,
		ShmSize:    req.ShmSize,
		Tmpfs:      req.Tmpfs,
	}

	networkingConfig := &network.NetworkingConfig{}

	// default hooks include logger hook and pre-create hook
	defaultHooks := []LifecycleHooks{
		DefaultLoggingHook(req.Logger),
		defaultPreCreateHook(dockerInput, hostConfig, networkingConfig),
		defaultCopyFileToContainerHook(req.Files),
		defaultLogConsumersHook(req.LogConsumerCfg),
		defaultReadinessHook(),
	}

	// in the case the container needs to access a local port
	// we need to forward the local port to the container
	if len(req.HostAccessPorts) > 0 {
		// a container lifecycle hook will be added, which will expose the host ports to the container
		// using a SSHD server running in a container. The SSHD server will be started and will
		// forward the host ports to the container ports.
		sshdForwardPortsHook, err := exposeHostPorts(ctx, &req, req.HostAccessPorts...)
		if err != nil {
			return nil, fmt.Errorf("failed to expose host ports: %w", err)
		}

		defaultHooks = append(defaultHooks, sshdForwardPortsHook)
	}

	req.LifecycleHooks = []LifecycleHooks{combineContainerHooks(defaultHooks, req.LifecycleHooks)}

	err = req.creatingHook(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := corecontainer.Create(ctx, dockerInput, hostConfig, networkingConfig, platform, req.Name)
	if err != nil {
		return nil, err
	}

	// #248: If there is more than one network specified in the request attach newly created container to them one by one
	if len(req.Networks) > 1 {
		for _, n := range req.Networks[1:] {
			nw, err := corenetwork.GetByName(ctx, n)
			if err == nil {
				endpointSetting := network.EndpointSettings{
					Aliases: req.NetworkAliases[n],
				}
				err = cli.NetworkConnect(ctx, nw.ID, resp.ID, &endpointSetting)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	c := &DockerContainer{
		ID:                resp.ID,
		WaitingFor:        req.WaitingFor,
		Image:             imageName,
		imageWasBuilt:     req.ShouldBuildImage(),
		keepBuiltImage:    req.ShouldKeepBuiltImage(),
		sessionID:         core.SessionID(),
		terminationSignal: termSignal,
		logger:            req.Logger,
		lifecycleHooks:    req.LifecycleHooks,
	}

	err = c.createdHook(ctx)
	if err != nil {
		return nil, err
	}

	// Disable cleanup on success
	termSignal = nil

	return c, nil
}

func reuseOrCreateContainer(ctx context.Context, req Request) (*DockerContainer, error) {
	c, err := findContainerByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if c == nil {
		createdContainer, err := newContainer(ctx, req)
		if err == nil {
			return createdContainer, nil
		}
		if !CreateFailDueToNameConflictRegex.MatchString(err.Error()) {
			return nil, err
		}

		c, err = waitContainerCreation(ctx, req.Name)
		if err != nil {
			return nil, err
		}
	}

	sessionID := core.SessionID()

	var termSignal chan bool
	if !config.Read().RyukDisabled {
		_, err := NewReaper(context.Background(), core.SessionID())
		if err != nil {
			return nil, fmt.Errorf("failed to create reaper: %w", err)
		}

		termSignal, err = reaper.Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to reaper: %w", err)
		}
	}

	// default hooks include logger hook and pre-create hook
	defaultHooks := []LifecycleHooks{
		DefaultLoggingHook(req.Logger),
		defaultReadinessHook(),
		defaultLogConsumersHook(req.LogConsumerCfg),
	}

	dc := &DockerContainer{
		ID:                c.ID,
		WaitingFor:        req.WaitingFor,
		Image:             c.Image,
		sessionID:         sessionID,
		terminationSignal: termSignal,
		logger:            req.Logger,
		lifecycleHooks:    []LifecycleHooks{combineContainerHooks(defaultHooks, req.LifecycleHooks)},
	}

	err = dc.startedHook(ctx)
	if err != nil {
		return nil, err
	}

	dc.isRunning = true

	err = dc.readiedHook(ctx)
	if err != nil {
		return nil, err
	}

	return dc, nil
}

func waitContainerCreation(ctx context.Context, name string) (*types.Container, error) {
	var ctr *types.Container
	return ctr, backoff.Retry(func() error {
		c, err := findContainerByName(ctx, name)
		if err != nil {
			if !errdefs.IsNotFound(err) && core.IsPermanentClientError(err) {
				return backoff.Permanent(err)
			}
			return err
		}

		if c == nil {
			return fmt.Errorf("container %s not found", name)
		}

		ctr = c
		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
}
