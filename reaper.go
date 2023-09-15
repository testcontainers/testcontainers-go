package testcontainers

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/internal/testcontainerssession"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// Deprecated: it has been replaced by the internal testcontainersdocker.LabelLang
	TestcontainerLabel = "org.testcontainers.golang"
	// Deprecated: it has been replaced by the internal testcontainersdocker.LabelSessionID
	TestcontainerLabelSessionID = TestcontainerLabel + ".sessionId"
	// Deprecated: it has been replaced by the internal testcontainersdocker.LabelReaper
	TestcontainerLabelIsReaper = TestcontainerLabel + ".reaper"

	ReaperDefaultImage = "docker.io/testcontainers/ryuk:0.5.1"
)

var (
	reaperInstance *Reaper // We would like to create reaper only once
	reaperMutex    sync.Mutex
	reaperOnce     sync.Once
)

// ReaperProvider represents a provider for the reaper to run itself with
// The ContainerProvider interface should usually satisfy this as well, so it is pluggable
type ReaperProvider interface {
	RunContainer(ctx context.Context, req ContainerRequest) (Container, error)
	Config() TestcontainersConfig
}

// NewReaper creates a Reaper with a sessionID to identify containers and a provider to use
// Deprecated: it's not possible to create a reaper anymore.
func NewReaper(ctx context.Context, sessionID string, provider ReaperProvider, reaperImageName string) (*Reaper, error) {
	return reuseOrCreateReaper(ctx, provider, WithImageName(reaperImageName))
}

// lookUpReaperContainer returns a DockerContainer type with the reaper container in the case
// it's found in the running state, and including the labels for sessionID, reaper, and ryuk.
// It will perform a retry with exponential backoff to allow for the container to be started and
// avoid potential false negatives.
func lookUpReaperContainer(ctx context.Context) (*DockerContainer, error) {
	dockerClient, err := NewDockerClientWithOpts(ctx)
	if err != nil {
		return nil, err
	}
	defer dockerClient.Close()

	// the backoff will take at most 5 seconds to find the reaper container
	// doing each attempt every 100ms
	exp := backoff.NewExponentialBackOff()

	// we want random intervals between 100ms and 500ms for concurrent executions
	// to not be synchronized: it could be the case that multiple executions of this
	// function happen at the same time (specially when called from a different test
	// process execution), and we want to avoid that they all try to find the reaper
	// container at the same time.
	exp.InitialInterval = time.Duration(rand.Intn(5)*100) * time.Millisecond
	exp.RandomizationFactor = rand.Float64() * 0.5
	exp.Multiplier = rand.Float64() * 2.0
	exp.MaxInterval = 5.0 * time.Second  // max interval between attempts
	exp.MaxElapsedTime = 1 * time.Minute // max time to keep trying

	var reaperContainer *DockerContainer
	err = backoff.Retry(func() error {
		args := []filters.KeyValuePair{
			filters.Arg("label", fmt.Sprintf("%s=%s", testcontainersdocker.LabelSessionID, testcontainerssession.SessionID())),
			filters.Arg("label", fmt.Sprintf("%s=%t", testcontainersdocker.LabelReaper, true)),
			filters.Arg("label", fmt.Sprintf("%s=%t", testcontainersdocker.LabelRyuk, true)),
		}

		resp, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{
			All:     true,
			Filters: filters.NewArgs(args...),
		})
		if err != nil {
			return err
		}

		if len(resp) == 0 {
			// reaper container not found in the running state: do not look for it again
			return nil
		}

		if len(resp) > 1 {
			panic(fmt.Sprintf("not possible to have multiple reaper containers found for session ID %s", testcontainerssession.SessionID()))
		}

		r, err := containerFromDockerResponse(ctx, resp[0])
		if err != nil {
			return err
		}

		reaperContainer = r

		return nil
	}, backoff.WithContext(exp, ctx))

	if err != nil {
		return nil, err
	}

	return reaperContainer, nil
}

// reuseOrCreateReaper returns an existing Reaper instance if it exists and is running. Otherwise, a new Reaper instance
// will be created with a sessionID to identify containers in the same test session/program.
func reuseOrCreateReaper(ctx context.Context, provider ReaperProvider, opts ...ContainerOption) (*Reaper, error) {
	reaperMutex.Lock()
	defer reaperMutex.Unlock()

	var reaperErr error
	reaperOnce.Do(func() {
		reaperContainer, err := lookUpReaperContainer(context.Background())
		if err == nil && reaperContainer != nil {
			// The reaper container exists as a Docker container: re-use it
			endpoint, err := reaperContainer.PortEndpoint(ctx, "8080", "")
			if err != nil {
				reaperErr = err
				return
			}

			Logger.Printf("🔥 Reaper obtained from Docker for this test session %s", testcontainerssession.SessionID())
			reaperInstance, reaperErr = &Reaper{
				Provider:  provider,
				SessionID: testcontainerssession.SessionID(),
				Endpoint:  endpoint,
				container: reaperContainer,
			}, nil
			return
		}

		// the container is not found at the Docker level: create it for first time in this test session
		r, err := newReaper(ctx, provider, opts...)
		if err != nil {
			reaperErr = err
			return
		}

		reaperInstance, reaperErr = r, nil
	})
	if reaperErr != nil {
		reaperOnce = sync.Once{}
		return nil, reaperErr
	}

	return reaperInstance, nil
}

// newReaper creates a Reaper with a sessionID to identify containers and a provider to use
// Do not call this directly, use reuseOrCreateReaper instead
func newReaper(ctx context.Context, provider ReaperProvider, opts ...ContainerOption) (*Reaper, error) {
	dockerHostMount := testcontainersdocker.ExtractDockerSocket(ctx)

	sessionID := testcontainerssession.SessionID()

	reaper := &Reaper{
		Provider:  provider,
		SessionID: sessionID,
	}

	listeningPort := nat.Port("8080/tcp")

	tcConfig := provider.Config().Config

	reaperOpts := containerOptions{}

	for _, opt := range opts {
		opt(&reaperOpts)
	}

	req := ContainerRequest{
		Image:         reaperImage(reaperOpts.ImageName),
		ExposedPorts:  []string{string(listeningPort)},
		Labels:        testcontainersdocker.DefaultLabels(),
		Mounts:        Mounts(BindMount(dockerHostMount, "/var/run/docker.sock")),
		Privileged:    tcConfig.RyukPrivileged,
		WaitingFor:    wait.ForListeningPort(listeningPort),
		ReaperOptions: opts,
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.AutoRemove = true
			hc.NetworkMode = Bridge
		},
	}

	// keep backwards compatibility
	req.ReaperImage = req.Image

	// include reaper-specific labels to the reaper container
	req.Labels[testcontainersdocker.LabelReaper] = "true"
	req.Labels[testcontainersdocker.LabelRyuk] = "true"

	// Attach reaper container to a requested network if it is specified
	if p, ok := provider.(*DockerProvider); ok {
		req.Networks = append(req.Networks, p.DefaultNetwork)
	}

	c, err := provider.RunContainer(ctx, req)
	if err != nil {
		return nil, err
	}
	reaper.container = c

	endpoint, err := c.PortEndpoint(ctx, "8080", "")
	if err != nil {
		return nil, err
	}
	reaper.Endpoint = endpoint

	return reaper, nil
}

// Reaper is used to start a sidecar container that cleans up resources
type Reaper struct {
	Provider  ReaperProvider
	SessionID string
	Endpoint  string
	container Container
}

// Connect runs a goroutine which can be terminated by sending true into the returned channel
func (r *Reaper) Connect() (chan bool, error) {
	conn, err := net.DialTimeout("tcp", r.Endpoint, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("%w: Connecting to Ryuk on %s failed", err, r.Endpoint)
	}

	terminationSignal := make(chan bool)
	go func(conn net.Conn) {
		sock := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		defer conn.Close()

		labelFilters := []string{}
		for l, v := range testcontainersdocker.DefaultLabels() {
			labelFilters = append(labelFilters, fmt.Sprintf("label=%s=%s", l, v))
		}

		retryLimit := 3
		for retryLimit > 0 {
			retryLimit--

			if _, err := sock.WriteString(strings.Join(labelFilters, "&")); err != nil {
				continue
			}

			if _, err := sock.WriteString("\n"); err != nil {
				continue
			}

			if err := sock.Flush(); err != nil {
				continue
			}

			resp, err := sock.ReadString('\n')
			if err != nil {
				continue
			}

			if resp == "ACK\n" {
				break
			}
		}

		<-terminationSignal
	}(conn)
	return terminationSignal, nil
}

// Labels returns the container labels to use so that this Reaper cleans them up
// Deprecated: internally replaced by testcontainersdocker.DefaultLabels()
func (r *Reaper) Labels() map[string]string {
	return map[string]string{
		testcontainersdocker.LabelLang:      "go",
		testcontainersdocker.LabelSessionID: r.SessionID,
	}
}

func reaperImage(reaperImageName string) string {
	if reaperImageName == "" {
		return ReaperDefaultImage
	}
	return reaperImageName
}
