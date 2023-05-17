package testcontainers

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go/internal"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// Deprecated: it has been replaced by the internal testcontainersdocker.LabelLang
	TestcontainerLabel = "org.testcontainers.golang"
	// Deprecated: it has been replaced by the internal testcontainersdocker.LabelSessionID
	TestcontainerLabelSessionID = TestcontainerLabel + ".sessionId"
	// Deprecated: it has been replaced by the internal testcontainersdocker.LabelReaper
	TestcontainerLabelIsReaper = TestcontainerLabel + ".reaper"

	ReaperDefaultImage = "docker.io/testcontainers/ryuk:0.4.0"
)

var (
	reaperInstance *Reaper // We would like to create reaper only once
	mutex          sync.Mutex
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
	return reuseOrCreateReaper(ctx, sessionID, provider, WithImageName(reaperImageName))
}

// reuseOrCreateReaper returns an existing Reaper instance if it exists and is running. Otherwise, a new Reaper instance
// will be created with a sessionID to identify containers and a provider to use
func reuseOrCreateReaper(ctx context.Context, sessionID string, provider ReaperProvider, opts ...ContainerOption) (*Reaper, error) {
	mutex.Lock()
	defer mutex.Unlock()
	// If reaper already exists and healthy, re-use it
	if reaperInstance != nil {
		// Verify this instance is still running by checking state.
		// Can't use Container.IsRunning because the bool is not updated when Reaper is terminated
		state, err := reaperInstance.container.State(ctx)
		if err == nil && state.Running {
			return reaperInstance, nil
		}
	}

	r, err := newReaper(ctx, sessionID, provider, opts...)
	if err != nil {
		return nil, err
	}

	reaperInstance = r
	return reaperInstance, nil
}

// newReaper creates a Reaper with a sessionID to identify containers and a provider to use
// Should only be used internally and instead use reuseOrCreateReaper to prefer reusing an existing Reaper instance
func newReaper(ctx context.Context, sessionID string, provider ReaperProvider, opts ...ContainerOption) (*Reaper, error) {
	dockerHost := testcontainersdocker.ExtractDockerHost(ctx)

	reaper := &Reaper{
		Provider:  provider,
		SessionID: sessionID,
	}

	listeningPort := nat.Port("8080/tcp")

	tcConfig := provider.Config()

	reaperOpts := containerOptions{}

	for _, opt := range opts {
		opt(&reaperOpts)
	}

	req := ContainerRequest{
		Image:        reaperImage(reaperOpts.ImageName),
		ExposedPorts: []string{string(listeningPort)},
		Labels: map[string]string{
			TestcontainerLabelIsReaper:       "true",
			testcontainersdocker.LabelReaper: "true",
		},
		Mounts:        Mounts(BindMount(dockerHost, "/var/run/docker.sock")),
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
	for k, v := range reaper.Labels() {
		if k == TestcontainerLabelSessionID || k == testcontainersdocker.LabelSessionID {
			continue
		}
		req.Labels[k] = v
	}

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
		for l, v := range r.Labels() {
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
func (r *Reaper) Labels() map[string]string {
	return map[string]string{
		TestcontainerLabel:                  "true",
		TestcontainerLabelSessionID:         r.SessionID,
		testcontainersdocker.LabelLang:      "go",
		testcontainersdocker.LabelVersion:   internal.Version,
		testcontainersdocker.LabelSessionID: r.SessionID,
	}
}

func reaperImage(reaperImageName string) string {
	if reaperImageName == "" {
		return ReaperDefaultImage
	}
	return reaperImageName
}
